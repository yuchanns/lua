package lua

// Lua basic types. These constants correspond to the values
// returned by `lua_type` and related functions in the Lua C API.
// See: https://www.lua.org/manual/5.4/manual.html#lua_type
const (
	LUA_TNONE          = -1 // no value/type
	LUA_TNIL           = 0  // Lua nil
	LUA_TBOOLEAN       = 1  // boolean type
	LUA_TLIGHTUSERDATA = 2  // light userdata
	LUA_TNUMBER        = 3  // number
	LUA_TSTRING        = 4  // string
	LUA_TTABLE         = 5  // table
	LUA_TFUNCTION      = 6  // function
	LUA_TUSERDATA      = 7  // full userdata
	LUA_TTHREAD        = 8  // thread
	LUA_TPROTO         = 9  // (internal) proto
)

// LUA_MULTRET is used for nresults in Call and PCall to indicate
// that all results from the called function should be returned.
// See: https://www.lua.org/manual/5.4/manual.html#lua_call
const LUA_MULTRET = -1

// Lua pseudo-indices used for accessing special tables such as the registry.
// See: https://www.lua.org/manual/5.4/manual.html#4.3
const (
	LUA_REGISTRYINDEX = (-LUAI_MAXSTACK - 1000)
)

// Thread status codes returned by Lua operations (see: https://www.lua.org/manual/5.4/manual.html#4.4)
const (
	LUA_OK        = 0 // success
	LUA_YIELD     = 1 // yielded
	LUA_ERRRUN    = 2 // runtime error
	LUA_ERRSYNTAX = 3 // syntax error
	LUA_ERRMEM    = 4 // memory allocation error
	LUA_ERRERR    = 5 // error while running the message handler
)

// Maximum Lua stack size (used for registry index calculation).
// See: https://www.lua.org/manual/5.4/manual.html#lua_Constants
const LUAI_MAXSTACK = 1000000

// Lua comparison and arithmetic operator codes, for use with lua_arith, lua_compare etc.
// See: https://www.lua.org/manual/5.4/manual.html#lua_arith
const (
	// Arithmetic operators
	LUA_OPADD  = 0  // addition
	LUA_OPSUB  = 1  // subtraction
	LUA_OPMUL  = 2  // multiplication
	LUA_OPMOD  = 3  // modulo
	LUA_OPPOW  = 4  // exponentiation
	LUA_OPDIV  = 5  // float division
	LUA_OPIDIV = 6  // integer division
	LUA_OPBAND = 7  // bitwise AND
	LUA_OPBOR  = 8  // bitwise OR
	LUA_OPBXOR = 9  // bitwise XOR
	LUA_OPSHL  = 10 // bitwise shift left
	LUA_OPSHR  = 11 // bitwise shift right
	LUA_OPUNM  = 12 // unary minus
	LUA_OPBNOT = 13 // bitwise NOT

	// Comparison operators
	LUA_OPEQ = 0 // equal
	LUA_OPLT = 1 // less than
	LUA_OPLE = 2 // less or equal
)

const LUA_MINSTACK = 20
