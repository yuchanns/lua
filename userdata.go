package lua

import (
	"unsafe"
)

// NewUserData creates a new full userdata object of the given size, pushes it onto the stack, and returns its pointer.
// This object can store arbitrary Go data for use with Lua.
// See: https://www.lua.org/manual/5.4/manual.html#lua_newuserdatauv
func (s *State) NewUserData(size int) unsafe.Pointer {
	if s.Version() < 504 {
		return s.ffi.LuaNewuserdata(s.luaL, size)
	}
	return s.NewUserDataUv(size, 1)
}

// NewUserDataUv creates a new userdata with the given size and a specified number of user values.
// Available since Lua 5.4.
// See: https://www.lua.org/manual/5.4/manual.html#lua_newuserdatauv
func (s *State) NewUserDataUv(size, nuv int) unsafe.Pointer {
	return s.ffi.LuaNewuserdatauv(s.luaL, size, nuv)
}

// GetIUserValue gets the nth user value associated with the userdata at idx (1-based).
// The result is pushed onto the Lua stack and its type code is returned.
// Available since Lua 5.4.
// See: https://www.lua.org/manual/5.4/manual.html#lua_getiuservalue
func (s *State) GetIUserValue(idx, n int) int {
	return int(s.ffi.LuaGetiuservalue(s.luaL, idx, n))
}

// SetIUserValue sets the nth user value of the userdata at idx with the value at the top of the stack.
// Available since Lua 5.4.
// See: https://www.lua.org/manual/5.4/manual.html#lua_setiuservalue
func (s *State) SetIUserValue(idx, n int) {
	s.ffi.LuaSetiuservalue(s.luaL, idx, n)
}

// GetUserValue gets the first user value associated with the userdata at idx.
// For most userdata, only one user value is used.
// See: https://www.lua.org/manual/5.4/manual.html#lua_getiuservalue
func (s *State) GetUserValue(idx int) int {
	if s.Version() < 504 {
		return int(s.ffi.LuaGetuservalue(s.luaL, idx))
	}
	return s.GetIUserValue(idx, 1)
}

// SetUserValue sets the first user value of the userdata at idx.
// See: https://www.lua.org/manual/5.4/manual.html#lua_setiuservalue
func (s *State) SetUserValue(idx int) {
	if s.Version() < 504 {
		s.ffi.LuaSetuservalue(s.luaL, idx)
		return
	}
	s.ffi.LuaSetiuservalue(s.luaL, idx, 1)
}

// CheckUserData checks that the value at ud is a userdata of the type given by tname and returns its pointer.
// Raises an error if the type does not match.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_checkudata
func (s *State) CheckUserData(ud int, tname string) (ptr unsafe.Pointer, err error) {
	tptr, err := bytePtrFromString(tname)
	if err != nil {
		return
	}
	return s.ffi.LuaLCheckudata(s.luaL, ud, tptr), nil
}

// TestUserData tests whether the value at ud is a userdata of the type given by tname, returning its pointer or nil.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_testudata
func (s *State) TestUserData(ud int, tname string) (ptr unsafe.Pointer, err error) {
	tptr, err := bytePtrFromString(tname)
	if err != nil {
		return
	}
	return s.ffi.LuaLTestudata(s.luaL, ud, tptr), nil
}
