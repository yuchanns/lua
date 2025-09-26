package lua

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

// NewThread creates a new Lua thread (coroutine), pushes it onto the stack, and returns its State.
// See: https://www.lua.org/manual/5.4/manual.html#lua_newthread
func (s *State) NewThread() *State {
	return BuildState(luaLib.ffi.LuaNewthread(s.luaL))
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
	err = s.CheckError(luaLib.ffi.LuaClosethread(s.luaL, fromL))
	return
}

// ResetThread is equivalent to CloseThread with from being nil.
// Available since Lua 5.4.
// Deprecated: use CloseThread(nil) instead.
// See: https://www.lua.org/manual/5.4/manual.html#lua_resetthread
func (s *State) ResetThread() (err error) {
	err = s.CheckError(luaLib.ffi.LuaResetthread(s.luaL))
	return
}

// PushThread pushes the current thread onto the Lua stack. Returns true if it's the main thread.
// See: https://www.lua.org/manual/5.4/manual.html#lua_pushthread
func (s *State) PushThread() (isMain bool) {
	isMain = luaLib.ffi.LuaPushthread(s.luaL) == 1
	return
}

type KFunc func(*State, int, unsafe.Pointer) int

// YieldK yields nresults values from the current coroutine, using continuation k and context ctx for resumption.
// Due to the limitation of Purego, only a limited number of callbacks may be created in a single Go
// process, and any memory allocated for these callbacks is never released.
// See: https://www.lua.org/manual/5.4/manual.html#lua_yieldk
func (s *State) YieldK(nresults int, ctx unsafe.Pointer, k KFunc) (err error) {
	protectionMsg := "unwinding protection"
	defer func() {
		if m := recover(); m != nil {
			if msg, ok := m.(string); !ok || msg != protectionMsg {
				panic(m) // re-raise the panic if it's not our hack
			}
		}
	}()

	var kb uintptr
	if k != nil {
		kb = purego.NewCallback(func(L unsafe.Pointer, status int, ctx unsafe.Pointer) int {
			// Use panic instead of setjmp/longjmp to avoid issues with syscall frames
			defer panic(protectionMsg)

			return k(BuildState(L), status, ctx)
		})
	}

	status := luaLib.ffi.LuaYieldk(s.luaL, nresults, ctx, kb)
	if status != LUA_OK && status != LUA_YIELD {
		err = s.CheckError(status)
	}
	return
}

// Yield yields nresults values from the current coroutine (no continuation function).
// See: https://www.lua.org/manual/5.4/manual.html#lua_yield
func (s *State) Yield(nresults int) (err error) {
	return s.YieldK(nresults, nil, nil)
}

// Resume resumes the given Lua thread, passing narg arguments, and returns possible error and
// whether it yielded.
// Available since lua5.4: return the number of results
// See: https://www.lua.org/manual/5.4/manual.html#lua_resume
func (s *State) Resume(from *State, narg int) (nres int32, yield bool, err error) {
	var fromL unsafe.Pointer
	if from != nil {
		fromL = from.luaL
	}
	var status int
	if luaLib.ffi.version >= 504 {
		status = luaLib.ffi.LuaResume(s.luaL, fromL, narg, unsafe.Pointer(&nres))
	} else {
		status = luaLib.ffi.LuaResume503(s.luaL, fromL, narg)
	}
	yield = status == LUA_YIELD
	if status != LUA_OK && status != LUA_YIELD {
		err = s.CheckError(status)
	}
	return
}

// Status returns the status code of the thread (running, yielded, etc).
// See: https://www.lua.org/manual/5.4/manual.html#lua_status
func (s *State) Status() int {
	return luaLib.ffi.LuaStatus(s.luaL)
}

// IsYieldable reports whether the current Lua thread is yieldable.
// See: https://www.lua.org/manual/5.4/manual.html#lua_isyieldable
func (s *State) IsYieldable() bool {
	return luaLib.ffi.LuaIsyieldable(s.luaL) == 1
}

// ToThread returns the Lua thread at the given stack index as a State.
// See: https://www.lua.org/manual/5.4/manual.html#lua_tothread
func (s *State) ToThread(idx int) *State {
	return BuildState(luaLib.ffi.LuaTothread(s.luaL, idx))
}

// ToPointer returns the Lua value at the given stack index as an unsafe.Pointer.
func (s *State) ToPointer(idx int) unsafe.Pointer {
	return luaLib.ffi.LuaTopointer(s.luaL, idx)
}
