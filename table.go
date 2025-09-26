package lua

// CreateTable creates a new empty table and pushes it onto the stack.
// Narr and nrec are hints for the array and hash part sizes.
// See: https://www.lua.org/manual/5.4/manual.html#lua_createtable
func (s *State) CreateTable(narr, nrec int) {
	luaLib.ffi.LuaCreatetable(s.luaL, narr, nrec)
}

// GetTable retrieves a value in table at idx using the key at the top of the stack, and pushes the result.
// See: https://www.lua.org/manual/5.4/manual.html#lua_gettable
func (s *State) GetTable(idx int) int {
	return luaLib.ffi.LuaGettable(s.luaL, idx)
}

// SetTable sets a value in a table at idx using a key-value pair from the stack.
// See: https://www.lua.org/manual/5.4/manual.html#lua_settable
func (s *State) SetTable(idx int) {
	luaLib.ffi.LuaSettable(s.luaL, idx)
}

// GetField pushes onto the stack the value of the field k from the table at idx.
// Returns the type of the pushed value.
// See: https://www.lua.org/manual/5.4/manual.html#lua_getfield
func (s *State) GetField(idx int, k string) (typ int) {
	p, _ := bytePtrFromString(k)
	typ = int(luaLib.ffi.LuaGetfield(s.luaL, idx, p))
	return
}

// SetField sets the field k of the table at idx using a value from the stack.
// See: https://www.lua.org/manual/5.4/manual.html#lua_setfield
func (s *State) SetField(idx int, k string) {
	p, _ := bytePtrFromString(k)
	luaLib.ffi.LuaSetfield(s.luaL, idx, p)
}

// GetI pushes onto the stack the value n from the table at idx (uses integer key n).
// Returns the value's type.
// See: https://www.lua.org/manual/5.4/manual.html#lua_geti
func (s *State) GetI(idx int, n int64) int {
	return luaLib.ffi.LuaGeti(s.luaL, idx, n)
}

// SetI sets a value at index n in the table at idx, using the value on top of the stack.
// See: https://www.lua.org/manual/5.4/manual.html#lua_seti
func (s *State) SetI(idx int, n int64) {
	luaLib.ffi.LuaSeti(s.luaL, idx, n)
}

// NewTable pushes a new empty table onto the stack.
// See: https://www.lua.org/manual/5.4/manual.html#lua_newtable
func (s *State) NewTable() {
	luaLib.ffi.LuaCreatetable(s.luaL, 0, 0)
}

// RawGet does a raw (no metamethods) lookup in table at idx using key from stack top.
// See: https://www.lua.org/manual/5.4/manual.html#lua_rawget
func (s *State) RawGet(idx int) int {
	return int(luaLib.ffi.LuaRawget(s.luaL, idx))
}

// RawSet does a raw (no metamethods) table set, using a key/value from the stack.
// See: https://www.lua.org/manual/5.4/manual.html#lua_rawset
func (s *State) RawSet(idx int) {
	luaLib.ffi.LuaRawset(s.luaL, idx)
}

// RawGetI retrieves the entry with key n from the table at idx, ignoring metamethods.
// See: https://www.lua.org/manual/5.4/manual.html#lua_rawgeti
func (s *State) RawGetI(idx int, n int64) int {
	return int(luaLib.ffi.LuaRawgeti(s.luaL, idx, n))
}

// RawSetI sets the value with key n in the table at idx, ignoring metamethods.
// See: https://www.lua.org/manual/5.4/manual.html#lua_rawseti
func (s *State) RawSetI(idx int, n int64) {
	luaLib.ffi.LuaRawseti(s.luaL, idx, n)
}

// RawGetP retrieves a value from a table at idx using a light userdata as the key.
// UNSAFE: The caller must ensure pointer validity for the Lua state duration.
// See: https://www.lua.org/manual/5.4/manual.html#lua_rawgetp
func (s *State) RawGetP(idx int, ud any) (typ int) {
	p := toLightUserData(ud)
	typ = int(luaLib.ffi.LuaRawgetp(s.luaL, idx, p))
	return
}

// RawSetP stores a value in a table at idx using a light userdata key.
// UNSAFE: The caller must ensure pointer validity for the Lua state duration.
// See: https://www.lua.org/manual/5.4/manual.html#lua_rawsetp
func (s *State) RawSetP(idx int, ud any) {
	p := toLightUserData(ud)
	luaLib.ffi.LuaRawsetp(s.luaL, idx, p)
}

// Next pops a key from the stack, and pushes the next key-value pair from table at idx.
// Returns false if no more elements.
// See: https://www.lua.org/manual/5.4/manual.html#lua_next
func (s *State) Next(idx int) bool {
	return luaLib.ffi.LuaNext(s.luaL, idx) != 0
}

// GeIMetaTable retrieves the metatable of the value at the given index and pushes it onto the stack.
// See: https://www.lua.org/manual/5.4/manual.html#lua_getmetatable
func (s *State) GeIMetaTable(index int) int {
	return luaLib.ffi.LuaGetmetatable(s.luaL, index)
}

// SetIMetaTable sets the metatable for the value at the given index.
// See: https://www.lua.org/manual/5.4/manual.html#lua_setmetatable
func (s *State) SetIMetaTable(index int) int {
	return luaLib.ffi.LuaSetmetatable(s.luaL, index)
}

// NewMetaTable creates a new metatable with the given name and pushes it onto the stack.
// Returns true if the metatable already existed.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_newmetatable
func (s *State) NewMetaTable(tname string) (has bool) {
	p, _ := bytePtrFromString(tname)
	has = luaLib.ffi.LuaLNewmetatable(s.luaL, p) == 0
	return
}

// SetMetaTable sets the metatable of the value at the top of the stack to the named metatable.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_setmetatable
func (s *State) SetMetaTable(tname string) {
	p, _ := bytePtrFromString(tname)
	luaLib.ffi.LuaLSetmetatable(s.luaL, p)
}

// GetMetaTable retrieves the metatable associated with the given name from the registry.
// Returns the type of the metatable.
func (s *State) GetMetaTable(tname string) (typ int) {
	return s.GetField(LUA_REGISTRYINDEX, tname)
}

// GetMetaField pushes the named metafield of the given object onto the stack.
// Returns the type of the metafield.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_getmetafield
func (s *State) GetMetaField(obj int, e string) (typ int) {
	p, _ := bytePtrFromString(e)
	typ = int(luaLib.ffi.LuaLGetmetafield(s.luaL, obj, p))
	return
}

// CallMeta calls the named metamethod on the given object.
// Returns true if the metamethod exists and was called.
// See: https://www.lua.org/manual/5.4/manual.html#luaL_callmeta
func (s *State) CallMeta(obj int, e string) (has bool) {
	p, _ := bytePtrFromString(e)
	has = luaLib.ffi.LuaLCallmeta(s.luaL, obj, p) == 1
	return
}
