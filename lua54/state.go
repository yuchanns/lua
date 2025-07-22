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

	// SAFETY: refAlloc holds a reference to the allocator's user data
	// to prevent garbage collection while the Lua state is active.
	refAlloc unsafe.Pointer
}

func newState(ffi *ffi, o *stateOpt) (state *State) {
	var L unsafe.Pointer
	var refAlloc unsafe.Pointer
	if o != nil {
		refAlloc = o.userData
		L = ffi.LuaNewstate(o.alloc, o.userData)
	} else {
		L = ffi.LuaLNewstate()
	}

	return &State{
		ffi:  ffi,
		luaL: L,

		refAlloc: refAlloc,
	}
}

func (s *State) OpenLibs() {
	s.ffi.LuaLOpenlibs(s.luaL)
}

func (s *State) Close() {
	if s.luaL == nil {
		return
	}

	s.ffi.LuaClose(s.luaL)
	s.luaL = nil
	s.refAlloc = nil
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

func (s *State) CheckError(status int) error {
	if status == LUA_OK {
		return nil
	}
	msg := s.ToString(-1)
	s.Pop(1)
	return &Error{
		status:  status,
		message: msg,
	}
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
	err = s.CheckError(s.ffi.LuaLoad(s.luaL, reader, unsafe.Pointer(&r), cname, m))
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
	err = s.CheckError(s.ffi.LuaLLoadbufferx(s.luaL, bf, sz, b, m))
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
	err = s.CheckError(s.ffi.LuaLLoadstring(s.luaL, n))
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
	err = s.CheckError(s.ffi.LuaLLoadfilex(s.luaL, fname, m))
	return
}

func (s *State) LoadFile(filename string) (err error) {
	return s.LoadFilex(filename)
}

func (s *State) PCall(nargs, nresults, errfunc int) (err error) {
	return s.PCallK(nargs, nresults, errfunc, 0, NoOpKFunction)
}

func (s *State) PCallK(nargs, nresults, errfunc int, ctx int, k LuaKFunction) (err error) {
	err = s.CheckError(s.ffi.LuaPcallk(s.luaL, nargs, nresults, errfunc, ctx, k))
	return
}

func (s *State) Call(nargs, nresults int) {
	s.CallK(nargs, nresults, 0, NoOpKFunction)
}

func (s *State) CallK(nargs, nresults, ctx int, k LuaKFunction) {
	s.ffi.LuaCallk(s.luaL, nargs, nresults, ctx, k)
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

var NoOpKFunction LuaKFunction = func(_ unsafe.Pointer, _ int, _ int) int {
	return 0
}
