package lua

import (
	"unsafe"

	"go.yuchanns.xyz/lua/internal/tools"
)

func (s *State) IsNumber(idx int) bool {
	return s.ffi.LuaIsnumber(s.luaL, idx) != 0
}

func (s *State) IsString(idx int) bool {
	return s.ffi.LuaIsstring(s.luaL, idx) != 0
}

func (s *State) IsCFunction(idx int) bool {
	return s.ffi.LuaIscfunction(s.luaL, idx) != 0
}

func (s *State) IsInteger(idx int) bool {
	return s.ffi.LuaIsinteger(s.luaL, idx) != 0
}

func (s *State) IsUserData(idx int) bool {
	return s.ffi.LuaIsuserdata(s.luaL, idx) != 0
}

func (s *State) Type(idx int) int {
	return s.ffi.LuaType(s.luaL, idx)
}

func (s *State) TypeName(tp int) string {
	p := s.ffi.LuaTypename(s.luaL, tp)
	if p == nil {
		return ""
	}
	return tools.BytePtrToString(p)
}

func (s *State) IsFunction(idx int) bool {
	return s.Type(idx) == LUA_TFUNCTION
}

func (s *State) IsNil(idx int) bool {
	return s.Type(idx) == LUA_TNIL
}

func (s *State) IsBoolean(idx int) bool {
	return s.Type(idx) == LUA_TBOOLEAN
}

func (s *State) IsNone(idx int) bool {
	return s.Type(idx) == LUA_TNONE
}

func (s *State) IsNoneOrNil(idx int) bool {
	return s.Type(idx) <= 0
}

func (s *State) IsLightUserData(idx int) bool {
	return s.Type(idx) == LUA_TLIGHTUSERDATA
}

func (s *State) ToNumberx(idx int, isnum bool) float64 {
	var isNumber int
	if isnum {
		isNumber = 1
	}
	return s.ffi.LuaTonumberx(s.luaL, idx, unsafe.Pointer(&isNumber))
}

func (s *State) ToIntegerx(idx int, isnum bool) int64 {
	var isNumber int
	if isnum {
		isNumber = 1
	}
	return s.ffi.LuaTointegerx(s.luaL, idx, unsafe.Pointer(&isNumber))
}

func (s *State) ToLString(idx int, size *int) string {
	p := s.ffi.LuaTolstring(s.luaL, idx, unsafe.Pointer(size))
	if p == nil {
		return ""
	}
	return tools.BytePtrToString(p)
}

func (s *State) ToBoolean(idx int) bool {
	return s.ffi.LuaToboolean(s.luaL, idx) != 0
}

func (s *State) ToNumber(idx int) float64 {
	return s.ToNumberx(idx, false)
}

func (s *State) ToInteger(idx int) int64 {
	return s.ToIntegerx(idx, false)
}

func (s *State) ToString(idx int) string {
	return s.ToLString(idx, nil)
}

func (s *State) ToUserData(idx int) unsafe.Pointer {
	return s.ffi.LuaTouserdata(s.luaL, idx)
}

func (s *State) ToCFunction(idx int) unsafe.Pointer {
	return s.ffi.LuaTocfunction(s.luaL, idx)
}

func (s *State) RawLen(idx int) uint {
	return s.ffi.LuaRawlen(s.luaL, idx)
}

func (s *State) CheckNumber(idx int) float64 {
	return s.ffi.LuaLChecknumber(s.luaL, idx)
}

func (s *State) CheckInteger(idx int) int64 {
	return s.ffi.LuaLCheckinteger(s.luaL, idx)
}

func (s *State) CheckLString(idx int, size int) string {
	p := s.ffi.LuaLChecklstring(s.luaL, idx, unsafe.Pointer(&size))
	if p == nil {
		return ""
	}
	return tools.BytePtrToString(p)
}

func (s *State) CheckType(idx int, tp int) {
	s.ffi.LuaLChecktype(s.luaL, idx, tp)
}

func (s *State) CheckAny(idx int) {
	s.ffi.LuaLCheckany(s.luaL, idx)
}

func (s *State) OptNumber(idx int, def float64) float64 {
	return s.ffi.LuaLOptnumber(s.luaL, idx, def)
}

func (s *State) OptInteger(idx int, def int64) int64 {
	return s.ffi.LuaLOptinteger(s.luaL, idx, def)
}

func (s *State) OptLString(idx int, def string, size *int) (string, error) {
	d, err := tools.BytePtrFromString(def)
	if err != nil {
		return "", err
	}
	p := s.ffi.LuaLOptlstring(s.luaL, idx, d, unsafe.Pointer(size))
	return tools.BytePtrToString(p), nil
}
