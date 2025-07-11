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
