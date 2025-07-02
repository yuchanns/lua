package lua

import (
	"fmt"
	"unsafe"

	"go.yuchanns.xyz/lua/internal/tools"
)

type State struct {
	ffi *ffi

	L unsafe.Pointer
}

func newState(ffi *ffi) (state *State) {
	L := ffi.LuaLNewstate()
	ffi.LuaLOpenlibs(L)

	return &State{
		ffi: ffi,
		L:   L,
	}
}

func (s *State) Close() {
	if s.L == nil {
		return
	}

	defer tools.FreeLibrary(s.ffi.lib)

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

func (s *State) ToString(idx int) string {
	return s.ToLString(idx, nil)
}

func (s *State) ToLString(idx int, size unsafe.Pointer) string {
	p := s.ffi.LuaTolstring(s.L, idx, size)
	if p == nil {
		return ""
	}
	return tools.BytePtrToString(p)
}

func (s *State) SetGlobal(name string) (err error) {
	n, err := tools.BytePtrFromString(name)
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
	n, err := tools.BytePtrFromString(scode)
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
