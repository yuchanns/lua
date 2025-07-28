package lua

import (
	"unsafe"

	"go.yuchanns.xyz/lua/internal/tools"
)

// IsNumber returns true if the value at idx is a number or can be converted to a number.
// See: https://www.lua.org/manual/5.4/manual.html#lua_isnumber
func (s *State) IsNumber(idx int) bool {
	return s.ffi.LuaIsnumber(s.luaL, idx) != 0
}

// IsString returns true if the value at idx is a string or can be converted to a string.
// See: https://www.lua.org/manual/5.4/manual.html#lua_isstring
func (s *State) IsString(idx int) bool {
	return s.ffi.LuaIsstring(s.luaL, idx) != 0
}

// IsGoFunction returns true if the value at idx is a C function.
// See: https://www.lua.org/manual/5.4/manual.html#lua_iscfunction
func (s *State) IsGoFunction(idx int) bool {
	return s.ffi.LuaIscfunction(s.luaL, idx) != 0
}

// IsInteger returns true if the value at idx is an integer.
// See: https://www.lua.org/manual/5.4/manual.html#lua_isinteger
func (s *State) IsInteger(idx int) bool {
	return s.ffi.LuaIsinteger(s.luaL, idx) != 0
}

// IsUserData returns true if the value at idx is a userdata or full userdata.
// See: https://www.lua.org/manual/5.4/manual.html#lua_isuserdata
func (s *State) IsUserData(idx int) bool {
	return s.ffi.LuaIsuserdata(s.luaL, idx) != 0
}

// Type returns the type code of the value at idx.
// See: https://www.lua.org/manual/5.4/manual.html#lua_type
func (s *State) Type(idx int) int {
	return s.ffi.LuaType(s.luaL, idx)
}

// TypeName returns the name of the given type code.
// See: https://www.lua.org/manual/5.4/manual.html#lua_typename
func (s *State) TypeName(tp int) string {
	p := s.ffi.LuaTypename(s.luaL, tp)
	if p == nil {
		return ""
	}
	return tools.BytePtrToString(p)
}

// IsFunction reports whether the value at idx is a Lua function.
// See: https://www.lua.org/manual/5.4/manual.html#lua_type
func (s *State) IsFunction(idx int) bool {
	return s.Type(idx) == LUA_TFUNCTION
}

// IsNil reports whether the value at idx is nil.
// See: https://www.lua.org/manual/5.4/manual.html#lua_type
func (s *State) IsNil(idx int) bool {
	return s.Type(idx) == LUA_TNIL
}

// IsBoolean reports whether the value at idx is a boolean.
// See: https://www.lua.org/manual/5.4/manual.html#lua_type
func (s *State) IsBoolean(idx int) bool {
	return s.Type(idx) == LUA_TBOOLEAN
}

// IsNone reports whether the value at idx is LUA_TNONE (stack index is not valid).
// See: https://www.lua.org/manual/5.4/manual.html#lua_type
func (s *State) IsNone(idx int) bool {
	return s.Type(idx) == LUA_TNONE
}

// IsNoneOrNil reports whether the value at idx is LUA_TNONE or LUA_TNIL.
func (s *State) IsNoneOrNil(idx int) bool {
	return s.Type(idx) <= 0
}

// IsLightUserData returns true if the value at idx is light userdata.
// See: https://www.lua.org/manual/5.4/manual.html#lua_type
func (s *State) IsLightUserData(idx int) bool {
	return s.Type(idx) == LUA_TLIGHTUSERDATA
}

// ToNumberx converts the value at idx to a number (float64).
// If isnum is true, sets a flag if conversion succeeds.
// See: https://www.lua.org/manual/5.4/manual.html#lua_tonumberx
func (s *State) ToNumberx(idx int, isnum bool) float64 {
	var isNumber int
	if isnum {
		isNumber = 1
	}
	return s.ffi.LuaTonumberx(s.luaL, idx, unsafe.Pointer(&isNumber))
}

// ToIntegerx converts the value at idx to an integer (int64).
// If isnum is true, sets a flag if conversion succeeds.
// See: https://www.lua.org/manual/5.4/manual.html#lua_tointegerx
func (s *State) ToIntegerx(idx int, isnum bool) int64 {
	var isNumber int
	if isnum {
		isNumber = 1
	}
	return s.ffi.LuaTointegerx(s.luaL, idx, unsafe.Pointer(&isNumber))
}

// ToLString converts the value at idx to a string and optionally returns its length.
// See: https://www.lua.org/manual/5.4/manual.html#lua_tolstring
func (s *State) ToLString(idx int, size *int) string {
	p := s.ffi.LuaTolstring(s.luaL, idx, unsafe.Pointer(size))
	if p == nil {
		return ""
	}
	return tools.BytePtrToString(p)
}

// ToBoolean converts the Lua value at idx to a Go boolean.
// See: https://www.lua.org/manual/5.4/manual.html#lua_toboolean
func (s *State) ToBoolean(idx int) bool {
	return s.ffi.LuaToboolean(s.luaL, idx) != 0
}

// ToNumber converts the value at idx to a Lua number (float64, without extra flag).
func (s *State) ToNumber(idx int) float64 {
	return s.ToNumberx(idx, false)
}

// ToInteger converts the value at idx to a Lua integer (int64, without extra flag).
func (s *State) ToInteger(idx int) int64 {
	return s.ToIntegerx(idx, false)
}

// ToString converts the value at idx to a Go string, using the default Lua string conversion.
func (s *State) ToString(idx int) string {
	return s.ToLString(idx, nil)
}

// ToUserData returns the userdata pointer at idx, or nil if it's not userdata.
// See: https://www.lua.org/manual/5.4/manual.html#lua_touserdata
func (s *State) ToUserData(idx int) unsafe.Pointer {
	return s.ffi.LuaTouserdata(s.luaL, idx)
}

// ToGoFunction returns the Go representation of the Lua C function at idx, or nil if not a function.
// See: https://www.lua.org/manual/5.4/manual.html#lua_tocfunction
func (s *State) ToGoFunction(idx int) GoFunc {
	fnptr := s.ffi.LuaTocfunction(s.luaL, idx)
	if fnptr == nil {
		return nil
	}
	return func(L *State) int {
		cFunc := *(*LuaCFunction)(fnptr)
		return cFunc(L.luaL)
	}
}

// RawLen returns the length of value at idx (arrays, strings, tables).
// See: https://www.lua.org/manual/5.4/manual.html#lua_rawlen
func (s *State) RawLen(idx int) uint {
	return s.ffi.LuaRawlen(s.luaL, idx)
}

// CheckNumber checks whether the value at idx is a number and returns it.
// Raises an error if it is not a number.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_checknumber
func (s *State) CheckNumber(idx int) float64 {
	return s.ffi.LuaLChecknumber(s.luaL, idx)
}

// CheckInteger checks whether the value at idx is an integer and returns it.
// Raises an error if it is not.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_checkinteger
func (s *State) CheckInteger(idx int) int64 {
	return s.ffi.LuaLCheckinteger(s.luaL, idx)
}

func (s *State) CheckString(idx int) string {
	return s.CheckLString(idx, nil)
}

// CheckLString checks whether the value at idx is a string, optionally returns its length, and returns the Go string.
// Raises an error if not string.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_checklstring
func (s *State) CheckLString(idx int, size *int) string {
	var sz unsafe.Pointer
	if size != nil {
		sz = unsafe.Pointer(size)
	}
	return tools.BytePtrToString(s.ffi.LuaLChecklstring(s.luaL, idx, sz))
}

// CheckType checks whether the value at idx has the given type, raising error if not.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_checktype
func (s *State) CheckType(idx int, tp int) {
	s.ffi.LuaLChecktype(s.luaL, idx, tp)
}

// CheckAny checks that the value at idx is not none (must exist, any type), raises error if none.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_checkany
func (s *State) CheckAny(idx int) {
	s.ffi.LuaLCheckany(s.luaL, idx)
}

// OptNumber fetches an optional number arg at idx, or uses def if not present or not number.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_optnumber
func (s *State) OptNumber(idx int, def float64) float64 {
	return s.ffi.LuaLOptnumber(s.luaL, idx, def)
}

// OptInteger fetches an optional integer arg at idx, or uses def if not present or not integer.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_optinteger
func (s *State) OptInteger(idx int, def int64) int64 {
	return s.ffi.LuaLOptinteger(s.luaL, idx, def)
}

// OptLString fetches an optional string arg at idx, or uses def if not present or not string.
// Returns the Go string, or def.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_optlstring
func (s *State) OptLString(idx int, def string, size *int) (string, error) {
	d, err := tools.BytePtrFromString(def)
	if err != nil {
		return "", err
	}
	p := s.ffi.LuaLOptlstring(s.luaL, idx, d, unsafe.Pointer(size))
	return tools.BytePtrToString(p), nil
}
