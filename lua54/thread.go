package lua

import "unsafe"

// NewThread creates a new Lua thread (coroutine), pushes it onto the stack, and returns its State.
// See: https://www.lua.org/manual/5.4/manual.html#lua_newthread
func (s *State) NewThread() *State {
	L := s.ffi.LuaNewthread(s.luaL)

	return &State{
		ffi:  s.ffi,
		luaL: L,
	}
}

// CloseThread closes the specified Lua thread (or the currently running thread if from is nil).
// Returns an error if closing fails.
// See: https://www.lua.org/manual/5.4/manual.html#lua_closethread
func (s *State) CloseThread(from *State) (err error) {
	var fromL unsafe.Pointer
	if from != nil {
		fromL = from.luaL
	}
	err = s.CheckError(s.ffi.LuaClosethread(s.luaL, fromL))
	return
}

// ResetThread is equivalent to CloseThread with from being nil.
// Deprecated: use CloseThread(nil) instead.
// See: https://www.lua.org/manual/5.4/manual.html#lua_resetthread
func (s *State) ResetThread() (err error) {
	err = s.CheckError(s.ffi.LuaResetthread(s.luaL))
	return
}

// PushThread pushes the current thread onto the Lua stack. Returns true if it's the main thread.
// See: https://www.lua.org/manual/5.4/manual.html#lua_pushthread
func (s *State) PushThread() (isMain bool) {
	isMain = s.ffi.LuaPushthread(s.luaL) == 1
	return
}

type KFunc func(*State, int, unsafe.Pointer) int

// YieldK yields nresults values from the current coroutine, using continuation k and context ctx for resumption.
// See: https://www.lua.org/manual/5.4/manual.html#lua_yieldk
func (s *State) YieldK(nresults int, ctx unsafe.Pointer, k KFunc) (err error) {
	status := s.ffi.LuaYieldk(s.luaL, nresults, ctx, func(L unsafe.Pointer, status int, ctx unsafe.Pointer) int {
		state := &State{
			ffi:  s.ffi,
			luaL: L,
		}
		return k(state, status, ctx)
	})
	if status != LUA_OK && status != LUA_YIELD {
		err = s.CheckError(status)
	}
	return
}

// Yield yields nresults values from the current coroutine (no continuation function).
// See: https://www.lua.org/manual/5.4/manual.html#lua_yield
func (s *State) Yield(nresults int) (err error) {
	return s.YieldK(nresults, nil, NoOpKFunc)
}

// Resume resumes the given Lua thread, passing narg arguments, and returns the number of results plus possible error.
// See: https://www.lua.org/manual/5.4/manual.html#lua_resume
func (s *State) Resume(from *State, narg int) (nres int32, err error) {
	var fromL unsafe.Pointer
	if from != nil {
		fromL = from.luaL
	}
	status := s.ffi.LuaResume(s.luaL, fromL, narg, unsafe.Pointer(&nres))
	if status != LUA_OK && status != LUA_YIELD {
		err = s.CheckError(status)
	}
	return
}

// Status returns the status code of the thread (running, yielded, etc).
// See: https://www.lua.org/manual/5.4/manual.html#lua_status
func (s *State) Status() int {
	return s.ffi.LuaStatus(s.luaL)
}

// IsYieldable reports whether the current Lua thread is yieldable.
// See: https://www.lua.org/manual/5.4/manual.html#lua_isyieldable
func (s *State) IsYieldable() bool {
	return s.ffi.LuaIsyieldable(s.luaL) == 1
}

// ToThread returns the Lua thread at the given stack index as a State.
// See: https://www.lua.org/manual/5.4/manual.html#lua_tothread
func (s *State) ToThread(idx int) *State {
	L := s.ffi.LuaTothread(s.luaL, idx)
	return &State{
		ffi:  s.ffi,
		luaL: L,
	}
}
