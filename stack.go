package lua

import (
	"unsafe"
)

// RawEqual reports whether the values at the given indices are primitively equal (using Lua's raw equality).
// See: https://www.lua.org/manual/5.4/manual.html#lua_rawequal
func (s *State) RawEqual(idx1, idx2 int) bool {
	return s.ffi.LuaRawequal(s.luaL, idx1, idx2) != 0
}

// Compare compares two values at the given indices with the specified Lua comparison operation opcode.
// See: https://www.lua.org/manual/5.4/manual.html#lua_compare
func (s *State) Compare(idx1, idx2, op int) bool {
	return s.ffi.LuaCompare(s.luaL, idx1, idx2, op) != 0
}

// Arith performs the given Lua arithmetic operation using the provided opcode on the top values of the stack.
// See: https://www.lua.org/manual/5.4/manual.html#lua_arith
func (s *State) Arith(op int) {
	s.ffi.LuaArith(s.luaL, op)
}

// Concat concatenates the top n values from the stack and pushes the result.
// See: https://www.lua.org/manual/5.4/manual.html#lua_concat
func (s *State) Concat(n int) {
	s.ffi.LuaConcat(s.luaL, n)
}

// Len computes the length of the value at the given stack index and pushes the result.
// See: https://www.lua.org/manual/5.4/manual.html#lua_len
func (s *State) Len(idx int) {
	s.ffi.LuaLen(s.luaL, idx)
}

// AbsIndex converts a possibly negative stack index into an absolute one.
// See: https://www.lua.org/manual/5.4/manual.html#lua_absindex
func (s *State) AbsIndex(idx int) int {
	return s.ffi.LuaAbsindex(s.luaL, idx)
}

// GetTop returns the current top index of the stack (number of elements).
// See: https://www.lua.org/manual/5.4/manual.html#lua_gettop
func (s *State) GetTop() int {
	return s.ffi.LuaGettop(s.luaL)
}

// SetTop sets the stack top to the given index, popping or pushing as needed.
// See: https://www.lua.org/manual/5.4/manual.html#lua_settop
func (s *State) SetTop(idx int) {
	s.ffi.LuaSettop(s.luaL, idx)
}

// PushValue pushes a copy of the element at the given stack index onto the stack.
// See: https://www.lua.org/manual/5.4/manual.html#lua_pushvalue
func (s *State) PushValue(idx int) {
	s.ffi.LuaPushvalue(s.luaL, idx)
}

// Rotate performs a circular rotation of n elements at the given index.
// See: https://www.lua.org/manual/5.4/manual.html#lua_rotate
func (s *State) Rotate(idx, n int) {
	s.ffi.LuaRotate(s.luaL, idx, n)
}

// Copy copies the value at fromidx to toidx in the stack, overwriting the destination.
// See: https://www.lua.org/manual/5.4/manual.html#lua_copy
func (s *State) Copy(fromidx, toidx int) {
	s.ffi.LuaCopy(s.luaL, fromidx, toidx)
}

// CheckStack ensures there is space for at least sz more elements on the stack.
// See: https://www.lua.org/manual/5.4/manual.html#lua_checkstack
func (s *State) CheckStack(sz int) bool {
	return s.ffi.LuaCheckstack(s.luaL, sz) != 0
}

// CheckStackMsg ensures there is space for at least sz more elements on the stack, raising an error with message if not.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_checkstack
func (s *State) CheckStackMsg(sz int, msg string) {
	m, _ := bytePtrFromString(msg)
	s.ffi.LuaLCheckstack(s.luaL, sz, m)
}

// XMove moves n values between stacks of different Lua states.
// See: https://www.lua.org/manual/5.4/manual.html#lua_xmove
func (s *State) XMove(to *State, n int) {
	s.ffi.LuaXmove(s.luaL, to.luaL, n)
}

// Pop removes n values from the top of the stack, equivalent to SetTop(-n-1).
func (s *State) Pop(n int) {
	s.SetTop(-n - 1)
}

// Insert moves the top element into the given position by rotating.
func (s *State) Insert(idx int) {
	s.Rotate(idx, 1)
}

// Remove deletes the element at idx by rotating it to the top and popping it.
func (s *State) Remove(idx int) {
	s.Rotate(idx, -1)
	s.Pop(1)
}

// Replace overwrites the element at idx with the value at the top, then pops the top value.
func (s *State) Replace(idx int) {
	s.Copy(-1, idx)
	s.Pop(1)
}

// PushNil pushes a nil value onto the stack.
// See: https://www.lua.org/manual/5.4/manual.html#lua_pushnil
func (s *State) PushNil() {
	s.ffi.LuaPushnil(s.luaL)
}

// PushNumber pushes a float64 value onto the stack as a Lua number.
// See: https://www.lua.org/manual/5.4/manual.html#lua_pushnumber
func (s *State) PushNumber(n float64) {
	s.ffi.LuaPushnumber(s.luaL, n)
}

// PushInteger pushes an int64 value onto the stack as a Lua integer.
// See: https://www.lua.org/manual/5.4/manual.html#lua_pushinteger
func (s *State) PushInteger(n int64) {
	s.ffi.LuaPushinteger(s.luaL, n)
}

// PushLString pushes a given Go string onto the stack as a Lua string with explicit length.
// See: https://www.lua.org/manual/5.4/manual.html#lua_pushlstring
func (s *State) PushLString(sv string) (ret *byte) {
	p, _ := bytePtrFromString(sv)
	ret = s.ffi.LuaPushlstring(s.luaL, p, len(sv))
	return
}

// PushString pushes a null-terminated string as a Lua string onto the stack.
// See: https://www.lua.org/manual/5.4/manual.html#lua_pushstring
func (s *State) PushString(sv string) (ret *byte) {
	p, _ := bytePtrFromString(sv)
	ret = s.ffi.LuaPushstring(s.luaL, p)
	return
}

// PushGoClousure pushes a Go function as a Lua C closure with n upvalues onto the stack.
// Caution: upvalues are read from the stack.
// See: https://www.lua.org/manual/5.4/manual.html#lua_pushcclosure
func (s *State) PushGoClousure(f GoFunc, n int) {
	s.ffi.LuaPushgoclosure(s.luaL, func(L unsafe.Pointer) int {
		state := s.Clone(L)
		return f(state)
	}, n)
}

// GetUpValue retrieves the name of the n-th upvalue of a function at funcindex.
func (s *State) SetUpValue(funcindex int, n int) (name string) {
	namePtr := s.ffi.LuaSetupvalue(s.luaL, funcindex, n)
	if namePtr != nil {
		name = bytePtrToString(namePtr)
	}
	return
}

// GetUpValue retrieves the name of the n-th upvalue of a function at funcindex.
func (s *State) GetUpValue(funcindex int, n int) (name string) {
	namePtr := s.ffi.LuaGetupvalue(s.luaL, funcindex, n)
	if namePtr != nil {
		name = bytePtrToString(namePtr)
	}
	return
}

// UpValueIndex returns the index of the n-th upvalue of a function.
func (s *State) UpValueIndex(n int) int {
	return LUA_REGISTRYINDEX - n
}

// PushBoolean pushes a Go boolean onto the stack as a Lua boolean value.
// See: https://www.lua.org/manual/5.4/manual.html#lua_pushboolean
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
func (s *State) PushLightUserData(ud any) {
	p := toLightUserData(ud)
	s.ffi.LuaPushlightuserdata(s.luaL, p)
}

// PushGoFunction pushes a Go CFunc as a Lua C function with no upvalues.
// A Go function is not convertible once pushed onto the stack.
// Use `ToCFunction` to get the C function pointer which wraps the Go function.
// See: https://www.lua.org/manual/5.4/manual.html#lua_pushcfunction
func (s *State) PushGoFunction(f GoFunc) {
	s.PushGoClousure(f, 0)
}

// PushCFunction pushes a C function pointer as a Lua C closure with no upvalues.
// Typically used with C function pointers from ToCFunction
func (s *State) PushCFunction(f unsafe.Pointer) {
	s.PushCClousure(f, 0)
}

// PushCClousure pushes a C function pointer as a Lua C closure with n upvalues.
// Typically used with C function pointers from ToCFunction
func (s *State) PushCClousure(f unsafe.Pointer, n int) {
	s.ffi.LuaPushcclousure(s.luaL, f, n)
}
