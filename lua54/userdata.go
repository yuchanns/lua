package lua

import (
	"unsafe"

	"go.yuchanns.xyz/lua/internal/tools"
)

func (s *State) NewUserData(size int) unsafe.Pointer {
	return s.NewUserDataUv(size, 1)
}

func (s *State) NewUserDataUv(size, nuv int) unsafe.Pointer {
	return s.ffi.LuaNewuserdatauv(s.luaL, size, nuv)
}

func (s *State) GetIUserValue(idx, n int) int {
	return int(s.ffi.LuaGetiuservalue(s.luaL, idx, n))
}

func (s *State) SetIUserValue(idx, n int) {
	s.ffi.LuaSetiuservalue(s.luaL, idx, n)
}

func (s *State) GetUserValue(idx int) int {
	return s.GetIUserValue(idx, 1)
}

func (s *State) SetUserValue(idx int) {
	s.ffi.LuaSetiuservalue(s.luaL, idx, 1)
}

func (s *State) CheckUserData(ud int, tname string) (ptr unsafe.Pointer, err error) {
	tptr, err := tools.BytePtrFromString(tname)
	if err != nil {
		return
	}
	return s.ffi.LuaLCheckudata(s.luaL, ud, tptr), nil
}

func (s *State) TestUserData(ud int, tname string) (ptr unsafe.Pointer, err error) {
	tptr, err := tools.BytePtrFromString(tname)
	if err != nil {
		return
	}
	return s.ffi.LuaLTestudata(s.luaL, ud, tptr), nil
}
