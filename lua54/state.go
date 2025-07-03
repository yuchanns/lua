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

type CFunc func(*State) int

func (s *State) AtPanic(fn CFunc) (old CFunc) {
	panicf := func(L unsafe.Pointer) int {
		state := &State{
			ffi:  s.ffi,
			luaL: L,
		}
		return fn(state)
	}
	oldptr := s.ffi.LuaAtpanic(s.luaL, panicf)
	oldCfunc := *(*LuaCFunction)(unsafe.Pointer(&oldptr))
	return func(state *State) int {
		L := unsafe.Pointer(state)
		return oldCfunc(L)
	}
}

func (s *State) Version() float64 {
	return s.ffi.LuaVersion(s.luaL)
}

func (s *State) PopError() (err error) {
	msg := s.ToString(-1)
	err = fmt.Errorf("%s", msg)
	s.Pop(1)
	return
}

func (s *State) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	b, _ := tools.BytePtrFromString(msg)
	s.ffi.LuaLError(s.luaL, b)
	return
}

func (s *State) SetGlobal(name string) (err error) {
	n, err := tools.BytePtrFromString(name)
	if err != nil {
		return
	}
	s.ffi.LuaSetglobal(s.luaL, n)
	return
}

func (s *State) CheckNumber(idx int) float64 {
	return s.ffi.LuaLChecknumber(s.luaL, idx)
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

// TODO: use State instead of unsafe.Pointer
func (s *State) SetWarnf(fn LuaWarnFunction, ud unsafe.Pointer) {
	s.ffi.LuaSetwarnf(s.luaL, fn, ud)
}
