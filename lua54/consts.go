package lua

const (
	// Basic types
	LUA_TNONE          = -1
	LUA_TNIL           = 0
	LUA_TBOOLEAN       = 1
	LUA_TLIGHTUSERDATA = 2
	LUA_TNUMBER        = 3
	LUA_TSTRING        = 4
	LUA_TTABLE         = 5
	LUA_TFUNCTION      = 6
	LUA_TUSERDATA      = 7
	LUA_TTHREAD        = 8
	LUA_TPROTO         = 9
)

// Option for multiple returns in `PCall` and `Call`
const LUA_MULTRET = -1

// Pseudo-indices
const (
	LUA_REGISTRYINDEX = (-LUAI_MAXSTACK - 1000)
)

const (
	// thread status
	LUA_OK        = 0
	LUA_YIELD     = 1
	LUA_ERRRUN    = 2
	LUA_ERRSYNTAX = 3
	LUA_ERRMEM    = 4
	LUA_ERRERR    = 5
)

const LUAI_MAXSTACK = 1000000

const (
	// Comparison and arithmetic operators
	LUA_OPADD  = 0
	LUA_OPSUB  = 1
	LUA_OPMUL  = 2
	LUA_OPMOD  = 3
	LUA_OPPOW  = 4
	LUA_OPDIV  = 5
	LUA_OPIDIV = 6
	LUA_OPBAND = 7
	LUA_OPBOR  = 8
	LUA_OPBXOR = 9
	LUA_OPSHL  = 10
	LUA_OPSHR  = 11
	LUA_OPUNM  = 12
	LUA_OPBNOT = 13

	LUA_OPEQ = 0
	LUA_OPLT = 1
	LUA_OPLE = 2
)
