package lua

import (
	"unsafe"

	"go.yuchanns.xyz/lua/internal/tools"
)

func (s *State) AbsIndex(idx int) int {
	return s.ffi.LuaAbsindex(s.luaL, idx)
}

func (s *State) GetTop() int {
	return s.ffi.LuaGettop(s.luaL)
}

func (s *State) SetTop(idx int) {
	s.ffi.LuaSettop(s.luaL, idx)
}

func (s *State) PushValue(idx int) {
	s.ffi.LuaPushvalue(s.luaL, idx)
}

func (s *State) Rotate(idx, n int) {
	s.ffi.LuaRotate(s.luaL, idx, n)
}

func (s *State) Copy(fromidx, toidx int) {
	s.ffi.LuaCopy(s.luaL, fromidx, toidx)
}

func (s *State) CheckStack(sz int) bool {
	return s.ffi.LuaCheckstack(s.luaL, sz) != 0
}

func (s *State) XMove(to *State, n int) {
	s.ffi.LuaXmove(s.luaL, to.luaL, n)
}

func (s *State) Pop(n int) {
	s.SetTop(-n - 1)
}

func (s *State) Insert(idx int) {
	s.Rotate(idx, 1)
}

func (s *State) Remove(idx int) {
	s.Rotate(idx, -1)
	s.Pop(1)
}

func (s *State) Replace(idx int) {
	s.Copy(-1, idx)
	s.Pop(1)
}

func (s *State) PushNil() {
	s.ffi.LuaPushnil(s.luaL)
}

func (s *State) PushNumber(n float64) {
	s.ffi.LuaPushnumber(s.luaL, n)
}

func (s *State) PushInteger(n int64) {
	s.ffi.LuaPushinteger(s.luaL, n)
}

func (s *State) PushLString(sv string) (ret *byte, err error) {
	p, err := tools.BytePtrFromString(sv)
	if err != nil {
		return
	}
	ret = s.ffi.LuaPushlstring(s.luaL, p, len(sv))
	return
}

func (s *State) PushString(sv string) (ret *byte, err error) {
	p, err := tools.BytePtrFromString(sv)
	if err != nil {
		return
	}
	ret = s.ffi.LuaPushstring(s.luaL, p)
	return
}

func (s *State) PushCClousure(f CFunc, n int) {
	s.ffi.LuaPushcclousure(s.luaL, func(L unsafe.Pointer) int {
		state := &State{
			ffi:  s.ffi,
			luaL: L,
		}
		return f(state)
	}, n)
}

func (s *State) PushBoolean(b bool) int {
	var v int
	if b {
		v = 1
	}
	return s.ffi.LuaPushboolean(s.luaL, v)
}

// PushLightUserData pushes a light userdata onto the stack.
// UNSAFE: The userdata must be a pointer type, and it is the caller's responsibility to ensure
// that the pointer remains valid for the lifetime of the Lua state.
func (s *State) PushLightUserData(ud any) (err error) {
	p, err := tools.ToLightUserData(ud)
	if err != nil {
		return
	}
	s.ffi.LuaPushlightuserdata(s.luaL, p)
	return
}

func (s *State) PushCFunction(f CFunc) {
	s.PushCClousure(f, 0)
}
