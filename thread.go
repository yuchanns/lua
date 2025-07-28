package lua

import "unsafe"

// NewThread creates a new Lua thread (coroutine), pushes it onto the stack, and returns its State.
// See: https://www.lua.org/manual/5.4/manual.html#lua_newthread
func (s *State) NewThread() *State {
	L := s.ffi.LuaNewthread(s.luaL)

	return s.clone(L)
}

// CloseThread closes the specified Lua thread (or the currently running thread if from is nil).
// Returns an error if closing fails.
// Available since Lua 5.4.
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
// Available since Lua 5.4.
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
	protectionMsg := "unwinding protection"
	if s.unwindingProtection {
		defer func() {
			if m := recover(); m != nil {
				if msg, ok := m.(string); !ok || msg != protectionMsg {
					panic(m) // re-raise the panic if it's not our hack
				}
			}
		}()
	}

	status := s.ffi.LuaYieldk(s.luaL, nresults, ctx, func(L unsafe.Pointer, status int, ctx unsafe.Pointer) int {
		if s.unwindingProtection {
			// Use panic instead of setjmp/longjmp to avoid issues with syscall frames
			defer panic(protectionMsg)
		}

		state := s.clone(L)
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

// Resume resumes the given Lua thread, passing narg arguments, and returns possible error.
// Available since lua5.4: return the number of results
// See: https://www.lua.org/manual/5.4/manual.html#lua_resume
func (s *State) Resume(from *State, narg int) (nres int32, err error) {
	var fromL unsafe.Pointer
	if from != nil {
		fromL = from.luaL
	}
	var status int
	if s.ffi.version >= 504 {
		status = s.ffi.LuaResume(s.luaL, fromL, narg, unsafe.Pointer(&nres))
	} else {
		status = s.ffi.LuaResume503(s.luaL, fromL, narg)
	}
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
	return s.clone(L)
}

// ToPointer returns the Lua value at the given stack index as an unsafe.Pointer.
func (s *State) ToPointer(idx int) unsafe.Pointer {
	return s.ffi.LuaTopointer(s.luaL, idx)
}
