package lua

import (
	"reflect"
	"strconv"
	"strings"
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
// This struct provides Go bindings to the Lua C API using purego FFI.
// We use `ffi` tag to specify the function name and version requirements for each entry point.
// The version requirements are specified using tags like "gte=504" for Lua 5.4.
type ffi struct {
	lib uintptr

	// State manipulation
	LuaNewstate    func(f LuaAlloc, ud unsafe.Pointer) unsafe.Pointer `ffi:"lua_newstate,gte=504"`
	LuaClose       func(L unsafe.Pointer)                             `ffi:"lua_close"`
	LuaNewthread   func(L unsafe.Pointer) unsafe.Pointer              `ffi:"lua_newthread,gte=504"`
	LuaClosethread func(L unsafe.Pointer, from unsafe.Pointer) int    `ffi:"lua_closethread,gte=504"`
	LuaResetthread func(L unsafe.Pointer) int                         `ffi:"lua_resetthread,gte=504"`

	LuaAtpanic func(L unsafe.Pointer, panicf LuaCFunction) unsafe.Pointer `ffi:"lua_atpanic,gte=504"`

	LuaVersion func(L unsafe.Pointer) float64 `ffi:"lua_version,gte=504"`

	// Basic stack manipulation
	LuaAbsindex   func(L unsafe.Pointer, idx int) int        `ffi:"lua_absindex,gte=504"`
	LuaGettop     func(L unsafe.Pointer) int                 `ffi:"lua_gettop,gte=504"`
	LuaSettop     func(L unsafe.Pointer, idx int)            `ffi:"lua_settop,gte=504"`
	LuaPushvalue  func(L unsafe.Pointer, idx int)            `ffi:"lua_pushvalue,gte=504"`
	LuaRotate     func(L unsafe.Pointer, idx, n int)         `ffi:"lua_rotate,gte=504"`
	LuaCopy       func(L unsafe.Pointer, fromidx, toidx int) `ffi:"lua_copy,gte=504"`
	LuaCheckstack func(L unsafe.Pointer, sz int) int         `ffi:"lua_checkstack,gte=504"`
	LuaXmove      func(from, to unsafe.Pointer, n int)       `ffi:"lua_xmove,gte=504"`

	// Access functions
	LuaIsnumber    func(L unsafe.Pointer, idx int) int  `ffi:"lua_isnumber,gte=504"`
	LuaIsstring    func(L unsafe.Pointer, idx int) int  `ffi:"lua_isstring,gte=504"`
	LuaIscfunction func(L unsafe.Pointer, idx int) int  `ffi:"lua_iscfunction,gte=504"`
	LuaIsinteger   func(L unsafe.Pointer, idx int) int  `ffi:"lua_isinteger,gte=504"`
	LuaIsuserdata  func(L unsafe.Pointer, idx int) int  `ffi:"lua_isuserdata,gte=504"`
	LuaType        func(L unsafe.Pointer, idx int) int  `ffi:"lua_type,gte=504"`
	LuaTypename    func(L unsafe.Pointer, tp int) *byte `ffi:"lua_typename,gte=504"`

	LuaTonumberx   func(L unsafe.Pointer, idx int, isnum unsafe.Pointer) float64 `ffi:"lua_tonumberx,gte=504"`
	LuaTointegerx  func(L unsafe.Pointer, idx int, isnum unsafe.Pointer) int64   `ffi:"lua_tointegerx,gte=504"`
	LuaTolstring   func(L unsafe.Pointer, idx int, sz unsafe.Pointer) *byte      `ffi:"lua_tolstring,gte=504"`
	LuaToboolean   func(L unsafe.Pointer, idx int) int                           `ffi:"lua_toboolean,gte=504"`
	LuaRawlen      func(L unsafe.Pointer, idx int) uint                          `ffi:"lua_rawlen,gte=504"`
	LuaTocfunction func(L unsafe.Pointer, idx int) unsafe.Pointer                `ffi:"lua_tocfunction,gte=504"`
	LuaTouserdata  func(L unsafe.Pointer, idx int) unsafe.Pointer                `ffi:"lua_touserdata,gte=504"`
	LuaTothread    func(L unsafe.Pointer, idx int) unsafe.Pointer                `ffi:"lua_tothread,gte=504"`

	LuaRawequal func(L unsafe.Pointer, idx1 int, idx2 int) int         `ffi:"lua_rawequal,gte=504"`
	LuaCompare  func(L unsafe.Pointer, idx1 int, idx2 int, op int) int `ffi:"lua_compare,gte=504"`
	LuaArith    func(L unsafe.Pointer, op int)                         `ffi:"lua_arith,gte=504"`
	LuaConcat   func(L unsafe.Pointer, n int)                          `ffi:"lua_concat,gte=504"`
	LuaLen      func(L unsafe.Pointer, idx int)                        `ffi:"lua_len,gte=504"`

	// Push functions
	LuaPushnil           func(L unsafe.Pointer)                         `ffi:"lua_pushnil,gte=504"`
	LuaPushnumber        func(L unsafe.Pointer, n float64)              `ffi:"lua_pushnumber,gte=504"`
	LuaPushinteger       func(L unsafe.Pointer, n int64)                `ffi:"lua_pushinteger,gte=504"`
	LuaPushlstring       func(L unsafe.Pointer, s *byte, len int) *byte `ffi:"lua_pushlstring,gte=504"`
	LuaPushstring        func(L unsafe.Pointer, s *byte) *byte          `ffi:"lua_pushstring,gte=504"`
	LuaPushcclousure     func(L unsafe.Pointer, f LuaCFunction, n int)  `ffi:"lua_pushcclosure,gte=504"`
	LuaPushboolean       func(L unsafe.Pointer, b int) int              `ffi:"lua_pushboolean,gte=504"`
	LuaPushlightuserdata func(L unsafe.Pointer, p unsafe.Pointer)       `ffi:"lua_pushlightuserdata,gte=504"`
	LuaPushthread        func(L unsafe.Pointer) int                     `ffi:"lua_pushthread,gte=504"`

	// Table and field functions
	LuaCreatetable func(L unsafe.Pointer, narr, nrec int)         `ffi:"lua_createtable,gte=504"`
	LuaGettable    func(L unsafe.Pointer, idx int) int            `ffi:"lua_gettable,gte=504"`
	LuaSettable    func(L unsafe.Pointer, idx int)                `ffi:"lua_settable,gte=504"`
	LuaGetfield    func(L unsafe.Pointer, idx int, k *byte) int32 `ffi:"lua_getfield,gte=504"`
	LuaSetfield    func(L unsafe.Pointer, idx int, k *byte)       `ffi:"lua_setfield,gte=504"`
	LuaGeti        func(L unsafe.Pointer, idx int, n int64) int   `ffi:"lua_geti,gte=504"`
	LuaSeti        func(L unsafe.Pointer, idx int, n int64)       `ffi:"lua_seti,gte=504"`
	// Table raw functions
	LuaRawget  func(L unsafe.Pointer, idx int) int32                   `ffi:"lua_rawget,gte=504"`
	LuaRawset  func(L unsafe.Pointer, idx int)                         `ffi:"lua_rawset,gte=504"`
	LuaRawgeti func(L unsafe.Pointer, idx int, n int64) int32          `ffi:"lua_rawgeti,gte=504"`
	LuaRawseti func(L unsafe.Pointer, idx int, n int64)                `ffi:"lua_rawseti,gte=504"`
	LuaRawgetp func(L unsafe.Pointer, idx int, p unsafe.Pointer) int32 `ffi:"lua_rawgetp,gte=504"`
	LuaRawsetp func(L unsafe.Pointer, idx int, p unsafe.Pointer)       `ffi:"lua_rawsetp,gte=504"`
	LuaNext    func(L unsafe.Pointer, idx int) int                     `ffi:"lua_next,gte=504"`
	// Meta table functions
	LuaGetmetatable func(L unsafe.Pointer, objindex int) int `ffi:"lua_getmetatable,gte=504"`
	LuaSetmetatable func(L unsafe.Pointer, objindex int) int `ffi:"lua_setmetatable,gte=504"`

	// Userdata functions
	LuaNewuserdatauv func(L unsafe.Pointer, sz int, nuvlue int) unsafe.Pointer  `ffi:"lua_newuserdatauv,gte=504"`
	LuaGetiuservalue func(L unsafe.Pointer, idx int, n int) int32               `ffi:"lua_getiuservalue,gte=504"`
	LuaSetiuservalue func(L unsafe.Pointer, idx int, n int)                     `ffi:"lua_setiuservalue,gte=504"`
	LuaLCheckudata   func(L unsafe.Pointer, ud int, tname *byte) unsafe.Pointer `ffi:"luaL_checkudata,gte=504"`
	LuaLTestudata    func(L unsafe.Pointer, ud int, tname *byte) unsafe.Pointer `ffi:"luaL_testudata,gte=504"`

	LuaGetglobal func(L unsafe.Pointer, name *byte) int32                                                     `ffi:"lua_getglobal,gte=504"`
	LuaSetglobal func(L unsafe.Pointer, name *byte)                                                           `ffi:"lua_setglobal,gte=504"`
	LuaCallk     func(L unsafe.Pointer, nargs, nresults int, ctx unsafe.Pointer, k LuaKFunction)              `ffi:"lua_callk,gte=504"`
	LuaPcallk    func(L unsafe.Pointer, nargs, nresults, errfunc int, ctx unsafe.Pointer, k LuaKFunction) int `ffi:"lua_pcallk,gte=504"`
	LuaLoad      func(L unsafe.Pointer, reader LuaReader, dt unsafe.Pointer, chunkname *byte, mode *byte) int `ffi:"lua_load,gte=504"`

	LuaSetwarnf func(L unsafe.Pointer, warnf LuaWarnFunction, ud unsafe.Pointer) `ffi:"lua_setwarnf,gte=504"`

	// Coroutine functions
	LuaYieldk      func(L unsafe.Pointer, nresults int, ctx unsafe.Pointer, k LuaKFunction) int   `ffi:"lua_yieldk,gte=504"`
	LuaResume      func(L unsafe.Pointer, from unsafe.Pointer, narg int, nres unsafe.Pointer) int `ffi:"lua_resume,gte=504"`
	LuaStatus      func(L unsafe.Pointer) int                                                     `ffi:"lua_status,gte=504"`
	LuaIsyieldable func(L unsafe.Pointer) int                                                     `ffi:"lua_isyieldable,gte=504"`

	LuaLNewstate func() unsafe.Pointer `ffi:"luaL_newstate"`
	// Open all preloaded libraries.
	LuaLOpenlibs func(L unsafe.Pointer) `ffi:"luaL_openlibs"`

	LuaLNewmetatable func(L unsafe.Pointer, tname *byte) int        `ffi:"luaL_newmetatable,gte=504"`
	LuaLSetmetatable func(L unsafe.Pointer, tname *byte)            `ffi:"luaL_setmetatable,gte=504"`
	LuaLCallmeta     func(L unsafe.Pointer, ojbj int, e *byte) int  `ffi:"luaL_callmeta,gte=504"`
	LuaLGetmetafield func(L unsafe.Pointer, obj int, e *byte) int32 `ffi:"luaL_getmetafield,gte=504"`

	// Auxiliary functions
	LuaLChecknumber  func(L unsafe.Pointer, idx int) float64                             `ffi:"luaL_checknumber,gte=504"`
	LuaLCheckinteger func(L unsafe.Pointer, idx int) int64                               `ffi:"luaL_checkinteger,gte=504"`
	LuaLChecklstring func(L unsafe.Pointer, idx int, sz unsafe.Pointer) *byte            `ffi:"luaL_checklstring,gte=504"`
	LuaLChecktype    func(L unsafe.Pointer, idx int, t int)                              `ffi:"luaL_checktype,gte=504"`
	LuaLCheckany     func(L unsafe.Pointer, idx int)                                     `ffi:"luaL_checkany,gte=504"`
	LuaLOptnumber    func(L unsafe.Pointer, idx int, def float64) float64                `ffi:"luaL_optnumber,gte=504"`
	LuaLOptinteger   func(L unsafe.Pointer, idx int, def int64) int64                    `ffi:"luaL_optinteger,gte=504"`
	LuaLOptlstring   func(L unsafe.Pointer, idx int, def *byte, sz unsafe.Pointer) *byte `ffi:"luaL_optlstring,gte=504"`
	LuaLCheckstack   func(L unsafe.Pointer, sz int, msg *byte) int                       `ffi:"luaL_checkstack,gte=504"`
	LuaLTolstring    func(L unsafe.Pointer, idx int, sz unsafe.Pointer) *byte            `ffi:"luaL_tolstring,gte=504"`

	LuaLError       func(L unsafe.Pointer, msg *byte) int                                  `ffi:"luaL_error,gte=504"`
	LuaLLoadstring  func(L unsafe.Pointer, s *byte) int                                    `ffi:"luaL_loadstring,gte=504"`
	LuaLLoadfilex   func(L unsafe.Pointer, filename *byte, mode *byte) int                 `ffi:"luaL_loadfilex,gte=504"`
	LuaLLoadbufferx func(L unsafe.Pointer, buff *byte, sz int, name *byte, mode *byte) int `ffi:"luaL_loadbufferx,gte=504"`
}

// getLuaVersion retrieves the Lua version from the loaded library.
// It uses the lua_version function to determine the version number.
// With this, we can conditionally register functions based on the Lua version.
func getLuaVersion(lib uintptr) (version int) {
	var (
		luaVersion   func(L *unsafe.Pointer) float64
		luaLNewState func() *unsafe.Pointer
		luaClose     func(L *unsafe.Pointer)
	)
	purego.RegisterLibFunc(&luaVersion, lib, "lua_version")
	purego.RegisterLibFunc(&luaLNewState, lib, "luaL_newstate")
	purego.RegisterLibFunc(&luaClose, lib, "lua_close")
	L := luaLNewState()
	defer luaClose(L)

	version = int(luaVersion(L))
	return
}

// newFFI loads the Lua 5.4 dynamic library at the specified path and registers all available exported entrypoints.
// It provides a Go ffi struct ready for low-level Lua C API interaction in memory.
func newFFI(path string) (FFI *ffi, err error) {
	lib, err := tools.LoadLibrary(path)
	if err != nil {
		return
	}

	version := getLuaVersion(lib)

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
		tag := field.Tag.Get("ffi")
		if tag == "" {
			continue
		}
		tags := strings.Split(tag, ",")
		if len(tags) == 0 {
			continue
		}
		fname := tags[0]
		var register = true
		for _, tag := range tags[1:] {
			tags := strings.Split(tag, "=")
			if len(tags) != 2 {
				continue
			}
			targetVersion, _ := strconv.Atoi(tags[1])
			switch tags[0] {
			case "gte":
				if version < targetVersion {
					register = false
					break
				}
			case "lte":
				if version > targetVersion {
					register = false
					break
				}
			}
		}

		if !register {
			continue
		}

		fptr := v.Field(i).Addr().Interface()

		purego.RegisterLibFunc(fptr, lib, fname)
	}
	return
}
