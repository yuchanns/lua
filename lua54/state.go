package lua

import (
	"fmt"
	"io"
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

type CFunc func(L *State) int

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

func (s *State) Errorf(format string, args ...any) {
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

// Load loads a Lua chunk without running it.
func (s *State) Load(r io.Reader, chunkname string, mode ...string) (err error) {
	cname, err := tools.BytePtrFromString(chunkname)
	if err != nil {
		return
	}
	var m *byte
	if len(mode) > 0 {
		m, err = tools.BytePtrFromString(mode[0])
		if err != nil {
			return
		}
	}

	buf := make([]byte, 4096)
	var reader LuaReader = func(_ unsafe.Pointer, ud unsafe.Pointer, sz *int) *byte {
		reader := *(*io.Reader)(ud)
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return nil
		}
		if n == 0 {
			return nil
		}
		*sz = n
		return &buf[0]
	}

	// SAFETY: it is safe to pass the reader as an unsafe.Pointer because
	// the reader immediately consumes the data from the io.Reader after
	// the call to LuaLoad. So it will not outlive the io.Reader.
	status := s.ffi.LuaLoad(s.luaL, reader, unsafe.Pointer(&r), cname, m)
	if status != LUA_OK {
		err = s.PopError()
	}
	return
}

func (s *State) LoadBuffer(buff []byte, name string) (err error) {
	return s.LoadBufferx(buff, name)
}

func (s *State) LoadBufferx(buff []byte, name string, mode ...string) (err error) {
	b, err := tools.BytePtrFromString(name)
	if err != nil {
		return
	}
	var m *byte
	if len(mode) > 0 {
		m, err = tools.BytePtrFromString(mode[0])
		if err != nil {
			return
		}
	}
	var bf *byte
	var sz = len(buff)
	if sz > 0 {
		bf = &buff[0]
	}
	status := s.ffi.LuaLLoadbufferx(s.luaL, bf, sz, b, m)
	if status != LUA_OK {
		err = s.PopError()
	}
	return
}

func (s *State) DoString(scode string) (err error) {
	err = s.LoadString(scode)
	if err != nil {
		return
	}
	return s.PCall(0, LUA_MULTRET, 0)
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

func (s *State) DoFile(filename string) (err error) {
	err = s.LoadFile(filename)
	if err != nil {
		return
	}
	return s.PCall(0, LUA_MULTRET, 0)
}

func (s *State) LoadFilex(filename string, mode ...string) (err error) {
	fname, err := tools.BytePtrFromString(filename)
	if err != nil {
		return
	}
	var m *byte
	if len(mode) > 0 {
		m, err = tools.BytePtrFromString(mode[0])
		if err != nil {
			return
		}
	}
	status := s.ffi.LuaLLoadfilex(s.luaL, fname, m)
	if status != LUA_OK {
		err = s.PopError()
	}
	return
}

func (s *State) LoadFile(filename string) (err error) {
	return s.LoadFilex(filename)
}

func (s *State) PCall(nargs, nresults, errfunc int) (err error) {
	status := s.ffi.LuaPcallk(s.luaL, nargs, nresults, errfunc, 0, noOpKFunction)
	if status != LUA_OK {
		err = s.PopError()
	}
	return
}

type WarnFunc func(L *State, msg string, tocont int)

func (s *State) SetWarnf(fn WarnFunc, ud unsafe.Pointer) {
	s.ffi.LuaSetwarnf(s.luaL, func(ud unsafe.Pointer, msg *byte, tocont int) {
		state := &State{
			ffi:  s.ffi,
			luaL: ud,
		}
		fn(state, tools.BytePtrToString(msg), tocont)
	}, ud)
}
