package lua

import (
	"go.yuchanns.xyz/lua/internal/tools"
)

func (s *State) CreateTable(narr, nrec int) {
	s.ffi.LuaCreatetable(s.luaL, narr, nrec)
}

func (s *State) GetTable(idx int) int {
	return s.ffi.LuaGettable(s.luaL, idx)
}

func (s *State) SetTable(idx int) {
	s.ffi.LuaSettable(s.luaL, idx)
}

func (s *State) GetField(idx int, k string) (typ int, err error) {
	p, err := tools.BytePtrFromString(k)
	if err != nil {
		return
	}
	typ = s.ffi.LuaGetfield(s.luaL, idx, p)
	return
}

func (s *State) SetField(idx int, k string) (err error) {
	p, err := tools.BytePtrFromString(k)
	if err != nil {
		return
	}
	s.ffi.LuaSetfield(s.luaL, idx, p)
	return
}

func (s *State) GetI(idx int, n int64) int {
	return s.ffi.LuaGeti(s.luaL, idx, n)
}

func (s *State) SetI(idx int, n int64) {
	s.ffi.LuaSeti(s.luaL, idx, n)
}

func (s *State) NewTable() {
	s.ffi.LuaCreatetable(s.luaL, 0, 0)
}

func (s *State) RawGet(idx int) int {
	return s.ffi.LuaRawget(s.luaL, idx)
}

func (s *State) RawSet(idx int) {
	s.ffi.LuaRawset(s.luaL, idx)
}

func (s *State) RawGetI(idx int, n int64) int {
	return s.ffi.LuaRawgeti(s.luaL, idx, n)
}

func (s *State) RawSetI(idx int, n int64) {
	s.ffi.LuaRawseti(s.luaL, idx, n)
}

// RawGetP retrieves a value from the stack at the given index using a light userdata pointer.
// UNSAFE: It is the caller's responsibility to ensure that the pointer remains valid for the
// lifetime of the Lua state.
func (s *State) RawGetP(idx int, ud any) (typ int, err error) {
	p, err := tools.ToLightUserData(ud)
	if err != nil {
		return
	}
	typ = s.ffi.LuaRawgetp(s.luaL, idx, p)
	return
}

// RawSetP sets a value at the given index using a light userdata pointer.
// UNSAFE: It is the caller's responsibility to ensure that the pointer remains valid for the
// lifetime of the Lua state.
func (s *State) RawSetP(idx int, ud any) (err error) {
	p, err := tools.ToLightUserData(ud)
	if err != nil {
		return
	}
	s.ffi.LuaRawsetp(s.luaL, idx, p)
	return
}

func (s *State) Next(idx int) bool {
	return s.ffi.LuaNext(s.luaL, idx) != 0
}
