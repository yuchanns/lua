package lua

import (
	"fmt"
	"reflect"
	"unsafe"
)

type State struct {
	ffi *ffi

	L unsafe.Pointer
}

func NewState(path string) (state *State, err error) {
	ffi, err := newFFI(path)
	if err != nil {
		return
	}

	L := ffi.LuaLNewstate()

	ffi.LuaLOpenlibs(L)

	state = &State{
		ffi: ffi,
		L:   L,
	}

	return
}

func (s *State) Close() {
	if s.L == nil {
		return
	}

	defer FreeLibrary(s.ffi.lib)

	s.ffi.LuaClose(s.L)
	s.L = nil
}

func (s *State) PopError() (err error) {
	msg := s.ToString(-1)
	err = fmt.Errorf("%s", msg)
	s.Pop(1)
	return
}

func (s *State) PushCClousure(f LuaCFunction, n int) {
	s.ffi.LuaPushcclousure(s.L, f, n)
}

func (s *State) PushCFunction(f LuaCFunction) {
	s.PushCClousure(f, 0)
}

func (s *State) PushGoFunction(f any) {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("expected a function, got %s", t.Kind()))
	}
	v := reflect.ValueOf(f)

	var fn LuaCFunction = func(L unsafe.Pointer) int {
		args := make([]reflect.Value, 0, t.NumIn())

		for i := range t.NumIn() {
			var arg reflect.Value
			switch t.In(i).Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				arg = reflect.ValueOf(s.ffi.LuaLCheckinteger(L, i+1))
			case reflect.Float64, reflect.Float32:
				arg = reflect.ValueOf(s.ffi.LuaLChecknumber(L, i+1))
			case reflect.String:
				p := s.ffi.LuaLChecklstring(L, i+1, nil)
				arg = reflect.ValueOf(BytePtrToString(p))
			case reflect.Bool:
				s.ffi.LuaLChecktype(L, i+1, LUA_TBOOLEAN)
				arg = reflect.ValueOf(s.ffi.LuaToboolean(L, i+1) != 0)
			default:
				panic(fmt.Sprintf("unsupported argument type: %s", t.In(i).Kind()))
			}
			args = append(args, arg)
		}
		results := v.Call(args)
		n := len(results)
		for i := range n {
			switch results[i].Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				s.ffi.LuaPushinteger(L, results[i].Int())
			case reflect.Float64, reflect.Float32:
				s.ffi.LuaPushnumber(L, results[i].Float())
			case reflect.String:
				str := results[i].String()
				nptr, err := BytePtrFromString(str)
				if err != nil {
					panic(fmt.Sprintf("failed to convert string to byte pointer: %v", err))
				}
				s.ffi.LuaPushlstring(L, nptr, len(str))
			case reflect.Bool:
				var b int
				if results[i].Bool() {
					b = 1
				}
				s.ffi.LuaPushboolean(L, b)
			default:
				panic(fmt.Sprintf("unsupported return type: %s", results[i].Kind()))
			}
		}

		return n
	}

	s.PushCFunction(fn)

	return
}

func (s *State) ToString(idx int) string {
	return s.ToLString(idx, nil)
}

func (s *State) ToLString(idx int, size unsafe.Pointer) string {
	p := s.ffi.LuaTolstring(s.L, idx, size)
	if p == nil {
		return ""
	}
	return BytePtrToString(p)
}

func (s *State) SetGlobal(name string) (err error) {
	n, err := BytePtrFromString(name)
	if err != nil {
		return
	}
	s.ffi.LuaSetglobal(s.L, n)
	return
}

func (s *State) GetTop() int {
	return s.ffi.LuaGettop(s.L)
}

func (s *State) SetTop(idx int) {
	s.ffi.LuaSettop(s.L, idx)
}

func (s *State) Pop(n int) {
	s.SetTop(-n - 1)
}

func (s *State) CheckNumber(idx int) float64 {
	return s.ffi.LuaLChecknumber(s.L, idx)
}

func (s *State) PushNumber(n float64) {
	s.ffi.LuaPushnumber(s.L, n)
}

func (s *State) DoString(scode string) (err error) {
	err = s.LoadString(scode)
	if err != nil {
		return
	}
	return s.PCall(0, 0, 0)
}

func (s *State) LoadString(scode string) (err error) {
	n, err := BytePtrFromString(scode)
	if err != nil {
		return
	}
	status := s.ffi.LuaLLoadstring(s.L, n)
	if status != LUA_OK {
		err = s.PopError()
	}
	return
}

func (s *State) PCall(nargs, nresults, errfunc int) (err error) {
	status := s.ffi.LuaPcallk(s.L, nargs, nresults, errfunc, 0, func(L unsafe.Pointer, status, ctx int) int {
		return 1
	})
	if status != LUA_OK {
		err = s.PopError()
	}
	return
}

func (s *State) SetWarnf(fn LuaWarnFunction, ud unsafe.Pointer) {
	s.ffi.LuaSetwarnf(s.L, fn, ud)
}
