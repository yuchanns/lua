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

	withoutUwindingProtection bool
}

// State represents a single Lua interpreter state, holding runtime and memory context.
// It is the Go binding for the Lua C API's lua_State pointer, supporting all standard C API operations.
// See: https://www.lua.org/manual/5.4/manual.html#lua_State
type State struct {
	ffi *ffi

	luaL unsafe.Pointer

	// SAFETY: refAlloc holds a reference to the allocator's user data
	// to prevent garbage collection while the Lua state is active.
	refAlloc unsafe.Pointer

	unwindingProtection bool
}

func newState(ffi *ffi, o *stateOpt) (L *State) {
	var luaL unsafe.Pointer
	var refAlloc unsafe.Pointer
	if o.userData != nil && o.alloc != nil {
		refAlloc = o.userData
		luaL = ffi.LuaNewstate(o.alloc, o.userData)
	} else {
		luaL = ffi.LuaLNewstate()
	}

	L = &State{
		ffi:  ffi,
		luaL: luaL,

		refAlloc:            refAlloc,
		unwindingProtection: !o.withoutUwindingProtection,
	}

	if L.unwindingProtection {
		// Convert Lua errors into Go panics
		L.AtPanic(func(L *State) int {
			err := L.checkUnprotectedError()

			panic(err)
		})
	}

	return L
}

// OpenLibs loads all standard Lua libraries into the current state.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_openlibs
func (s *State) OpenLibs() {
	s.ffi.LuaLOpenlibs(s.luaL)
}

// Close properly shuts down and deallocates the Lua state, freeing any owned resources.
// After calling Close, the State must not be used again.
// See: https://www.lua.org/manual/5.4/manual.html#lua_close
func (s *State) Close() {
	if s.luaL == nil {
		return
	}

	s.ffi.LuaClose(s.luaL)
	s.luaL = nil
	s.refAlloc = nil
}

type GoFunc func(L *State) int

// AtPanic sets a Go function as the Lua panic handler for this state, returning the old panic handler.
// See: https://www.lua.org/manual/5.4/manual.html#lua_atpanic
func (s *State) AtPanic(fn GoFunc) (old GoFunc) {
	panicf := func(L unsafe.Pointer) int {
		state := s.clone(L)
		return fn(state)
	}
	oldptr := s.ffi.LuaAtpanic(s.luaL, panicf)
	oldCfunc := *(*LuaCFunction)(unsafe.Pointer(&oldptr))
	return func(state *State) int {
		L := unsafe.Pointer(state)
		return oldCfunc(L)
	}
}

// Version returns the current version of the Lua runtime loaded in this state.
// See: https://www.lua.org/manual/5.4/manual.html#lua_version
func (s *State) Version() float64 {
	return s.ffi.version
}

// CheckError transforms a Lua C API error code into a Go error,
// automatically extracting the human-readable message from the stack if needed.
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

func (s *State) clone(L unsafe.Pointer) *State {
	return &State{
		ffi:                 s.ffi,
		luaL:                L,
		refAlloc:            s.refAlloc,
		unwindingProtection: s.unwindingProtection,
	}
}

func (s *State) checkUnprotectedError() error {
	msg := s.ToString(-1)
	s.Pop(1)
	return &UnprotectedError{message: msg}
}

// Errorf raises a formatted Lua error from the Go side, pushing the error onto the Lua stack.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_error
func (s *State) Errorf(format string, args ...any) int {
	msg := fmt.Sprintf(format, args...)
	b, _ := tools.BytePtrFromString(msg)
	return s.ffi.LuaLError(s.luaL, b)
}

// SetGlobal sets a global variable in the Lua environment using the value at the top of the stack.
// See: https://www.lua.org/manual/5.4/manual.html#lua_setglobal
func (s *State) SetGlobal(name string) (err error) {
	n, err := tools.BytePtrFromString(name)
	if err != nil {
		return
	}
	s.ffi.LuaSetglobal(s.luaL, n)
	return
}

// GetGlobal retrieves a global variable from the Lua environment and pushes it onto the stack.
// See: https://www.lua.org/manual/5.4/manual.html#lua_getglobal
func (s *State) GetGlobal(name string) (err error) {
	n, err := tools.BytePtrFromString(name)
	if err != nil {
		return
	}
	s.ffi.LuaGetglobal(s.luaL, n)
	return
}

// Load loads a Lua chunk from an io.Reader, compiling but not executing the code. This mirrors lua_load.
// See: https://www.lua.org/manual/5.4/manual.html#lua_load
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

// LoadBuffer loads a Lua chunk from a byte slice with the given chunk name.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_loadbuffer
func (s *State) LoadBuffer(buff []byte, name string) (err error) {
	return s.LoadBufferx(buff, name)
}

// LoadBufferx is the extended form of LoadBuffer supporting the mode parameter, as in luaL_loadbufferx.
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

// DoString loads and runs a given Lua string in the current state. Returns any error encountered.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_dostring
func (s *State) DoString(scode string) (err error) {
	err = s.LoadString(scode)
	if err != nil {
		return
	}
	return s.PCall(0, LUA_MULTRET, 0)
}

// LoadString loads a Lua chunk from a Go string with the provided source code.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_loadstring
func (s *State) LoadString(scode string) (err error) {
	n, err := tools.BytePtrFromString(scode)
	if err != nil {
		return
	}
	err = s.CheckError(s.ffi.LuaLLoadstring(s.luaL, n))
	return
}

// DoFile loads and runs a Lua source file.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_dofile
func (s *State) DoFile(filename string) (err error) {
	err = s.LoadFile(filename)
	if err != nil {
		return
	}
	return s.PCall(0, LUA_MULTRET, 0)
}

// LoadFilex loads (but does not run) a Lua source file, optionally specifying the mode (text, binary, or both).
// See: https://www.lua.org/manual/5.4/manual.html#luaL_loadfilex
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

// LoadFile loads a Lua source file from disk without executing it.
func (s *State) LoadFile(filename string) (err error) {
	return s.LoadFilex(filename)
}

// PCall calls a Lua function in protected mode, with argument and result counts. If errors occur, they are returned.
// See: https://www.lua.org/manual/5.4/manual.html#lua_pcall
func (s *State) PCall(nargs, nresults, errfunc int) (err error) {
	return s.PCallK(nargs, nresults, errfunc, nil, NoOpKFunction)
}

// PCallK is like PCall but with full support for Lua continuation functions and execution contexts. Used for advanced coroutine yield/resume situations.
// See: https://www.lua.org/manual/5.4/manual.html#lua_pcallk
func (s *State) PCallK(nargs, nresults, errfunc int, ctx unsafe.Pointer, k LuaKFunction) (err error) {
	if !s.unwindingProtection {
		return s.CheckError(s.ffi.LuaPcallk(s.luaL, nargs, nresults, errfunc, ctx, k))
	}
	defer func() {
		if r := recover(); r != nil {
			err = &Error{
				status:  LUA_ERRRUN,
				message: fmt.Sprintf("%v", r),
			}
		}
	}()
	s.CallK(nargs, nresults, ctx, k)
	return
}

// Call invokes a Lua function (not in protected mode) with given arg and result counts. Panics on error.
// See: https://www.lua.org/manual/5.4/manual.html#lua_call
func (s *State) Call(nargs, nresults int) {
	s.CallK(nargs, nresults, nil, NoOpKFunction)
}

// CallK calls a Lua function with the given continuation and context, supporting advanced coroutine control.
func (s *State) CallK(nargs, nresults int, ctx unsafe.Pointer, k LuaKFunction) {
	s.ffi.LuaCallk(s.luaL, nargs, nresults, ctx, k)
}

type WarnFunc func(L *State, msg string, tocont int)

// SetWarnf sets a Go warning callback for this Lua state, called on warnings/errors from the Lua VM.
// See: https://www.lua.org/manual/5.4/manual.html#lua_setwarnf
func (s *State) SetWarnf(fn WarnFunc, ud unsafe.Pointer) {
	s.ffi.LuaSetwarnf(s.luaL, func(ud unsafe.Pointer, msg *byte, tocont int) {
		state := s.clone(ud)
		fn(state, tools.BytePtrToString(msg), tocont)
	}, ud)
}

var NoOpKFunction LuaKFunction = func(_ unsafe.Pointer, _ int, _ unsafe.Pointer) int {
	return 0
}

var NoOpKFunc KFunc = func(_ *State, _ int, _ unsafe.Pointer) int {
	return 0
}

// Requiref loads a Lua module by name, calling the provided Go function to open it.
func (s *State) Requiref(modname string, openf GoFunc, global bool) (err error) {
	mname, err := tools.BytePtrFromString(modname)
	if err != nil {
		return
	}
	var glb int
	if global {
		glb = 1
	}
	s.ffi.LuaLRequiref(s.luaL, mname, func(L unsafe.Pointer) int {
		state := s.clone(L)
		return openf(state)
	}, glb)
	return
}

// Ref creates a reference to the value at the given stack index, returning a unique reference ID.
func (s *State) Ref(idx int) int {
	return s.ffi.LuaLRef(s.luaL, idx)
}

// Unref removes a reference created by Ref, the entry is removed from the table.
func (s *State) Unref(idx int, ref int) {
	s.ffi.LuaLUnref(s.luaL, idx, ref)
}
