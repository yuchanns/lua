package lua

import (
	"reflect"
	"unsafe"

	"github.com/ebitengine/purego"
	"go.yuchanns.xyz/lua/internal/tools"
)

// LuaReader represents the Go equivalent of the lua_Reader C callback type for streaming data into the Lua state.
// See: https://www.lua.org/manual/5.4/manual.html#lua_Reader
type LuaReader func(L unsafe.Pointer, ud unsafe.Pointer, sz *int) *byte

// LuaAlloc mirrors the lua_Alloc C function type, used for advanced state memory allocation customization.
// See: https://www.lua.org/manual/5.4/manual.html#lua_Alloc
type LuaAlloc func(ud unsafe.Pointer, ptr unsafe.Pointer, osize, nsize int) unsafe.Pointer

// LuaCFunction is the Go equivalent of the C lua_CFunction for stack-based callbacks.
// See: https://www.lua.org/manual/5.4/manual.html#lua_CFunction
type LuaCFunction func(L unsafe.Pointer) int

// LuaKFunction is the Go equivalent for lua_KFunction, supporting continuation-style yields from C to Lua.
// See: https://www.lua.org/manual/5.4/manual.html#lua_KFunction
type LuaKFunction func(L unsafe.Pointer, status int, ctx unsafe.Pointer) int

// LuaWarnFunction is a Go representation of the Lua C API lua_WarnFunction for error and warning hooks.
// See: https://www.lua.org/manual/5.4/manual.html#lua_WarnFunction
type LuaWarnFunction func(ud unsafe.Pointer, msg *byte, tocont int)

// ffi stores all dynamically loaded Lua C API entry points for runtime use.
// This struct provides Go bindings to the Lua 5.4 C API using purego FFI.
type ffi struct {
	lib uintptr

	// State manipulation
	LuaNewstate    func(f LuaAlloc, ud unsafe.Pointer) unsafe.Pointer `ffi:"lua_newstate"`
	LuaClose       func(L unsafe.Pointer)                             `ffi:"lua_close"`
	LuaNewthread   func(L unsafe.Pointer) unsafe.Pointer              `ffi:"lua_newthread"`
	LuaClosethread func(L unsafe.Pointer, from unsafe.Pointer) int    `ffi:"lua_closethread"`
	LuaResetthread func(L unsafe.Pointer) int                         `ffi:"lua_resetthread"`

	LuaAtpanic func(L unsafe.Pointer, panicf LuaCFunction) unsafe.Pointer `ffi:"lua_atpanic"`

	LuaVersion func(L unsafe.Pointer) float64 `ffi:"lua_version"`

	// Basic stack manipulation
	LuaAbsindex   func(L unsafe.Pointer, idx int) int        `ffi:"lua_absindex"`
	LuaGettop     func(L unsafe.Pointer) int                 `ffi:"lua_gettop"`
	LuaSettop     func(L unsafe.Pointer, idx int)            `ffi:"lua_settop"`
	LuaPushvalue  func(L unsafe.Pointer, idx int)            `ffi:"lua_pushvalue"`
	LuaRotate     func(L unsafe.Pointer, idx, n int)         `ffi:"lua_rotate"`
	LuaCopy       func(L unsafe.Pointer, fromidx, toidx int) `ffi:"lua_copy"`
	LuaCheckstack func(L unsafe.Pointer, sz int) int         `ffi:"lua_checkstack"`
	LuaXmove      func(from, to unsafe.Pointer, n int)       `ffi:"lua_xmove"`

	// Access functions
	LuaIsnumber    func(L unsafe.Pointer, idx int) int  `ffi:"lua_isnumber"`
	LuaIsstring    func(L unsafe.Pointer, idx int) int  `ffi:"lua_isstring"`
	LuaIscfunction func(L unsafe.Pointer, idx int) int  `ffi:"lua_iscfunction"`
	LuaIsinteger   func(L unsafe.Pointer, idx int) int  `ffi:"lua_isinteger"`
	LuaIsuserdata  func(L unsafe.Pointer, idx int) int  `ffi:"lua_isuserdata"`
	LuaType        func(L unsafe.Pointer, idx int) int  `ffi:"lua_type"`
	LuaTypename    func(L unsafe.Pointer, tp int) *byte `ffi:"lua_typename"`

	LuaTonumberx   func(L unsafe.Pointer, idx int, isnum unsafe.Pointer) float64 `ffi:"lua_tonumberx"`
	LuaTointegerx  func(L unsafe.Pointer, idx int, isnum unsafe.Pointer) int64   `ffi:"lua_tointegerx"`
	LuaTolstring   func(L unsafe.Pointer, idx int, sz unsafe.Pointer) *byte      `ffi:"lua_tolstring"`
	LuaToboolean   func(L unsafe.Pointer, idx int) int                           `ffi:"lua_toboolean"`
	LuaRawlen      func(L unsafe.Pointer, idx int) uint                          `ffi:"lua_rawlen"`
	LuaTocfunction func(L unsafe.Pointer, idx int) unsafe.Pointer                `ffi:"lua_tocfunction"`
	LuaTouserdata  func(L unsafe.Pointer, idx int) unsafe.Pointer                `ffi:"lua_touserdata"`
	LuaTothread    func(L unsafe.Pointer, idx int) unsafe.Pointer                `ffi:"lua_tothread"`

	LuaRawequal func(L unsafe.Pointer, idx1 int, idx2 int) int         `ffi:"lua_rawequal"`
	LuaCompare  func(L unsafe.Pointer, idx1 int, idx2 int, op int) int `ffi:"lua_compare"`
	LuaArith    func(L unsafe.Pointer, op int)                         `ffi:"lua_arith"`
	LuaConcat   func(L unsafe.Pointer, n int)                          `ffi:"lua_concat"`
	LuaLen      func(L unsafe.Pointer, idx int)                        `ffi:"lua_len"`

	// Push functions
	LuaPushnil           func(L unsafe.Pointer)                         `ffi:"lua_pushnil"`
	LuaPushnumber        func(L unsafe.Pointer, n float64)              `ffi:"lua_pushnumber"`
	LuaPushinteger       func(L unsafe.Pointer, n int64)                `ffi:"lua_pushinteger"`
	LuaPushlstring       func(L unsafe.Pointer, s *byte, len int) *byte `ffi:"lua_pushlstring"`
	LuaPushstring        func(L unsafe.Pointer, s *byte) *byte          `ffi:"lua_pushstring"`
	LuaPushcclousure     func(L unsafe.Pointer, f LuaCFunction, n int)  `ffi:"lua_pushcclosure"`
	LuaPushboolean       func(L unsafe.Pointer, b int) int              `ffi:"lua_pushboolean"`
	LuaPushlightuserdata func(L unsafe.Pointer, p unsafe.Pointer)       `ffi:"lua_pushlightuserdata"`
	LuaPushthread        func(L unsafe.Pointer) int                     `ffi:"lua_pushthread"`

	// Table and field functions
	LuaCreatetable func(L unsafe.Pointer, narr, nrec int)         `ffi:"lua_createtable"`
	LuaGettable    func(L unsafe.Pointer, idx int) int            `ffi:"lua_gettable"`
	LuaSettable    func(L unsafe.Pointer, idx int)                `ffi:"lua_settable"`
	LuaGetfield    func(L unsafe.Pointer, idx int, k *byte) int32 `ffi:"lua_getfield"`
	LuaSetfield    func(L unsafe.Pointer, idx int, k *byte)       `ffi:"lua_setfield"`
	LuaGeti        func(L unsafe.Pointer, idx int, n int64) int   `ffi:"lua_geti"`
	LuaSeti        func(L unsafe.Pointer, idx int, n int64)       `ffi:"lua_seti"`
	// Table raw functions
	LuaRawget  func(L unsafe.Pointer, idx int) int32                   `ffi:"lua_rawget"`
	LuaRawset  func(L unsafe.Pointer, idx int)                         `ffi:"lua_rawset"`
	LuaRawgeti func(L unsafe.Pointer, idx int, n int64) int32          `ffi:"lua_rawgeti"`
	LuaRawseti func(L unsafe.Pointer, idx int, n int64)                `ffi:"lua_rawseti"`
	LuaRawgetp func(L unsafe.Pointer, idx int, p unsafe.Pointer) int32 `ffi:"lua_rawgetp"`
	LuaRawsetp func(L unsafe.Pointer, idx int, p unsafe.Pointer)       `ffi:"lua_rawsetp"`
	LuaNext    func(L unsafe.Pointer, idx int) int                     `ffi:"lua_next"`
	// Meta table functions
	LuaGetmetatable func(L unsafe.Pointer, objindex int) int `ffi:"lua_getmetatable"`
	LuaSetmetatable func(L unsafe.Pointer, objindex int) int `ffi:"lua_setmetatable"`

	// Userdata functions
	LuaNewuserdatauv func(L unsafe.Pointer, sz int, nuvlue int) unsafe.Pointer  `ffi:"lua_newuserdatauv"`
	LuaGetiuservalue func(L unsafe.Pointer, idx int, n int) int32               `ffi:"lua_getiuservalue"`
	LuaSetiuservalue func(L unsafe.Pointer, idx int, n int)                     `ffi:"lua_setiuservalue"`
	LuaLCheckudata   func(L unsafe.Pointer, ud int, tname *byte) unsafe.Pointer `ffi:"luaL_checkudata"`
	LuaLTestudata    func(L unsafe.Pointer, ud int, tname *byte) unsafe.Pointer `ffi:"luaL_testudata"`

	LuaSetglobal func(L unsafe.Pointer, name *byte)                                                           `ffi:"lua_setglobal"`
	LuaCallk     func(L unsafe.Pointer, nargs, nresults int, ctx unsafe.Pointer, k LuaKFunction)              `ffi:"lua_callk"`
	LuaPcallk    func(L unsafe.Pointer, nargs, nresults, errfunc int, ctx unsafe.Pointer, k LuaKFunction) int `ffi:"lua_pcallk"`
	LuaLoad      func(L unsafe.Pointer, reader LuaReader, dt unsafe.Pointer, chunkname *byte, mode *byte) int `ffi:"lua_load"`

	LuaSetwarnf func(L unsafe.Pointer, warnf LuaWarnFunction, ud unsafe.Pointer) `ffi:"lua_setwarnf"`

	// Coroutine functions
	LuaYieldk      func(L unsafe.Pointer, nresults int, ctx unsafe.Pointer, k LuaKFunction) int   `ffi:"lua_yieldk"`
	LuaResume      func(L unsafe.Pointer, from unsafe.Pointer, narg int, nres unsafe.Pointer) int `ffi:"lua_resume"`
	LuaStatus      func(L unsafe.Pointer) int                                                     `ffi:"lua_status"`
	LuaIsyieldable func(L unsafe.Pointer) int                                                     `ffi:"lua_isyieldable"`

	LuaLNewstate func() unsafe.Pointer `ffi:"luaL_newstate"`
	// Open all preloaded libraries.
	LuaLOpenlibs func(L unsafe.Pointer) `ffi:"luaL_openlibs"`

	LuaLNewmetatable func(L unsafe.Pointer, tname *byte) int        `ffi:"luaL_newmetatable"`
	LuaLSetmetatable func(L unsafe.Pointer, tname *byte)            `ffi:"luaL_setmetatable"`
	LuaLCallmeta     func(L unsafe.Pointer, ojbj int, e *byte) int  `ffi:"luaL_callmeta"`
	LuaLGetmetafield func(L unsafe.Pointer, obj int, e *byte) int32 `ffi:"luaL_getmetafield"`

	// Auxiliary functions
	LuaLChecknumber  func(L unsafe.Pointer, idx int) float64                             `ffi:"luaL_checknumber"`
	LuaLCheckinteger func(L unsafe.Pointer, idx int) int64                               `ffi:"luaL_checkinteger"`
	LuaLChecklstring func(L unsafe.Pointer, idx int, sz unsafe.Pointer) *byte            `ffi:"luaL_checklstring"`
	LuaLChecktype    func(L unsafe.Pointer, idx int, t int)                              `ffi:"luaL_checktype"`
	LuaLCheckany     func(L unsafe.Pointer, idx int)                                     `ffi:"luaL_checkany"`
	LuaLOptnumber    func(L unsafe.Pointer, idx int, def float64) float64                `ffi:"luaL_optnumber"`
	LuaLOptinteger   func(L unsafe.Pointer, idx int, def int64) int64                    `ffi:"luaL_optinteger"`
	LuaLOptlstring   func(L unsafe.Pointer, idx int, def *byte, sz unsafe.Pointer) *byte `ffi:"luaL_optlstring"`
	LuaLCheckstack   func(L unsafe.Pointer, sz int, msg *byte) int                       `ffi:"luaL_checkstack"`
	LuaLTolstring    func(L unsafe.Pointer, idx int, sz unsafe.Pointer) *byte            `ffi:"luaL_tolstring"`

	LuaLError       func(L unsafe.Pointer, msg *byte) int                                  `ffi:"luaL_error"`
	LuaLLoadstring  func(L unsafe.Pointer, s *byte) int                                    `ffi:"luaL_loadstring"`
	LuaLLoadfilex   func(L unsafe.Pointer, filename *byte, mode *byte) int                 `ffi:"luaL_loadfilex"`
	LuaLLoadbufferx func(L unsafe.Pointer, buff *byte, sz int, name *byte, mode *byte) int `ffi:"luaL_loadbufferx"`
}

// newFFI loads the Lua 5.4 dynamic library at the specified path and registers all available exported entrypoints.
// It provides a Go ffi struct ready for low-level Lua C API interaction in memory.
func newFFI(path string) (FFI *ffi, err error) {
	lib, err := tools.LoadLibrary(path)
	if err != nil {
		return
	}

	FFI = &ffi{
		lib: lib,
	}

	t := reflect.TypeOf(FFI).Elem()
	v := reflect.ValueOf(FFI).Elem()

	for i := range t.NumField() {
		field := t.Field(i)
		if field.Type.Kind() != reflect.Func {
			continue
		}
		fname := field.Tag.Get("ffi")
		if fname == "" {
			continue
		}
		fptr := v.Field(i).Addr().Interface()

		purego.RegisterLibFunc(fptr, lib, fname)
	}
	return
}
