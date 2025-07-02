package lua

import (
	"fmt"
	"unsafe"

	"go.yuchanns.xyz/lua/internal/tools"
)

type stateOpt struct {
	alloc    LuaAlloc
	userData unsafe.Pointer
}

type State struct {
	ffi *ffi

	luaL unsafe.Pointer
}

func newState(ffi *ffi, o *stateOpt) (state *State) {
	var L unsafe.Pointer
	if o != nil {
		L = ffi.LuaNewstate(o.alloc, o.userData)
	} else {
		L = ffi.LuaLNewstate()
	}
	ffi.LuaLOpenlibs(L)

	return &State{
		ffi:  ffi,
		luaL: L,
	}
}

func (s *State) Close() {
	if s.luaL == nil {
		return
	}

	s.ffi.LuaClose(s.luaL)
	s.luaL = nil
}

func (s *State) PopError() (err error) {
	msg := s.ToString(-1)
	err = fmt.Errorf("%s", msg)
	s.Pop(1)
	return
}

func (s *State) PushCClousure(f LuaCFunction, n int) {
	s.ffi.LuaPushcclousure(s.luaL, f, n)
}

func (s *State) PushCFunction(f LuaCFunction) {
	s.PushCClousure(f, 0)
}

func (s *State) ToString(idx int) string {
	return s.ToLString(idx, nil)
}

func (s *State) ToLString(idx int, size unsafe.Pointer) string {
	p := s.ffi.LuaTolstring(s.luaL, idx, size)
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
	s.ffi.LuaSetglobal(s.luaL, n)
	return
}

func (s *State) GetTop() int {
	return s.ffi.LuaGettop(s.luaL)
}

func (s *State) SetTop(idx int) {
	s.ffi.LuaSettop(s.luaL, idx)
}

func (s *State) Pop(n int) {
	s.SetTop(-n - 1)
}

func (s *State) CheckNumber(idx int) float64 {
	return s.ffi.LuaLChecknumber(s.luaL, idx)
}

func (s *State) PushNumber(n float64) {
	s.ffi.LuaPushnumber(s.luaL, n)
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
	status := s.ffi.LuaLLoadstring(s.luaL, n)
	if status != LUA_OK {
		err = s.PopError()
	}
	return
}

func (s *State) PCall(nargs, nresults, errfunc int) (err error) {
	status := s.ffi.LuaPcallk(s.luaL, nargs, nresults, errfunc, 0, NoOpKFunction)
	if status != LUA_OK {
		err = s.PopError()
	}
	return
}

func (s *State) SetWarnf(fn LuaWarnFunction, ud unsafe.Pointer) {
	s.ffi.LuaSetwarnf(s.luaL, fn, ud)
}
