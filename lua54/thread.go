package lua

import "unsafe"

func (s *State) NewThread() *State {
	L := s.ffi.LuaNewthread(s.luaL)

	return &State{
		ffi:  s.ffi,
		luaL: L,
	}
}

func (s *State) CloseThread(from *State) (err error) {
	var fromL unsafe.Pointer
	if from != nil {
		fromL = from.luaL
	}
	err = s.CheckError(s.ffi.LuaClosethread(s.luaL, fromL))
	return
}

// ResetThread is equivalent to `CloseThread` with `from` being nil
// Deprecated: use `CloseThread(nil)` instead.
func (s *State) ResetThread() (err error) {
	err = s.CheckError(s.ffi.LuaResetthread(s.luaL))
	return
}

func (s *State) PushThread() (isMain bool) {
	isMain = s.ffi.LuaPushthread(s.luaL) == 1
	return
}

func (s *State) YieldK(nresults int, ctx int, k LuaKFunction) (err error) {
	status := s.ffi.LuaYieldk(s.luaL, nresults, ctx, k)
	if status != LUA_OK && status != LUA_YIELD {
		err = s.CheckError(status)
	}
	return
}

func (s *State) Yield(nresults int) (err error) {
	return s.YieldK(nresults, 0, NoOpKFunction)
}

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

func (s *State) Status() int {
	return s.ffi.LuaStatus(s.luaL)
}

func (s *State) IsYieldable() bool {
	return s.ffi.LuaIsyieldable(s.luaL) == 1
}

func (s *State) ToThread(idx int) *State {
	L := s.ffi.LuaTothread(s.luaL, idx)
	return &State{
		ffi:  s.ffi,
		luaL: L,
	}
}
