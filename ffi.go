package lua

import (
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/ebitengine/purego"
)

// LuaReader represents the Go equivalent of the lua_Reader C callback type for streaming data into the Lua state.
// See: https://www.lua.org/manual/5.4/manual.html#lua_Reader
type LuaReader func(L unsafe.Pointer, ud unsafe.Pointer, sz *int) *byte

// LuaKFunction is the Go equivalent for lua_KFunction, supporting continuation-style yields from C to Lua.
// See: https://www.lua.org/manual/5.4/manual.html#lua_KFunction
type LuaKFunction func(L unsafe.Pointer, status int, ctx unsafe.Pointer) int

type LuaLReg struct {
	Name *byte
	Func unsafe.Pointer
}

// ffi stores all dynamically loaded Lua C API entry points for runtime use.
// This struct provides Go bindings to the Lua C API using purego FFI.
// We use `ffi` tag to specify the function name and version requirements for each entry point.
// The version requirements are specified using tags like "gte=503" for Lua 5.4.
type ffi struct {
	lib     uintptr
	version float64

	// State manipulation
	LuaNewstate    func(f uintptr, ud unsafe.Pointer) unsafe.Pointer `ffi:"lua_newstate,gte=503"`
	LuaClose       func(L unsafe.Pointer)                            `ffi:"lua_close"`
	LuaNewthread   func(L unsafe.Pointer) unsafe.Pointer             `ffi:"lua_newthread,gte=503"`
	LuaClosethread func(L unsafe.Pointer, from unsafe.Pointer) int   `ffi:"lua_closethread,gte=504"`
	LuaResetthread func(L unsafe.Pointer) int                        `ffi:"lua_resetthread,gte=504"`

	LuaAtpanic func(L unsafe.Pointer, panicf uintptr) unsafe.Pointer `ffi:"lua_atpanic,gte=503"`

	LuaVersion func(L unsafe.Pointer) float64 `ffi:"lua_version,gte=503"`

	// Basic stack manipulation
	LuaAbsindex   func(L unsafe.Pointer, idx int) int        `ffi:"lua_absindex,gte=503"`
	LuaGettop     func(L unsafe.Pointer) int                 `ffi:"lua_gettop,gte=503"`
	LuaSettop     func(L unsafe.Pointer, idx int)            `ffi:"lua_settop,gte=503"`
	LuaPushvalue  func(L unsafe.Pointer, idx int)            `ffi:"lua_pushvalue,gte=503"`
	LuaRotate     func(L unsafe.Pointer, idx, n int)         `ffi:"lua_rotate,gte=503"`
	LuaCopy       func(L unsafe.Pointer, fromidx, toidx int) `ffi:"lua_copy,gte=503"`
	LuaCheckstack func(L unsafe.Pointer, sz int) int         `ffi:"lua_checkstack,gte=503"`
	LuaXmove      func(from, to unsafe.Pointer, n int)       `ffi:"lua_xmove,gte=503"`

	// Access functions
	LuaIsnumber    func(L unsafe.Pointer, idx int) int  `ffi:"lua_isnumber,gte=503"`
	LuaIsstring    func(L unsafe.Pointer, idx int) int  `ffi:"lua_isstring,gte=503"`
	LuaIscfunction func(L unsafe.Pointer, idx int) int  `ffi:"lua_iscfunction,gte=503"`
	LuaIsinteger   func(L unsafe.Pointer, idx int) int  `ffi:"lua_isinteger,gte=503"`
	LuaIsuserdata  func(L unsafe.Pointer, idx int) int  `ffi:"lua_isuserdata,gte=503"`
	LuaType        func(L unsafe.Pointer, idx int) int  `ffi:"lua_type,gte=503"`
	LuaTypename    func(L unsafe.Pointer, tp int) *byte `ffi:"lua_typename,gte=503"`

	LuaTonumberx   func(L unsafe.Pointer, idx int, isnum unsafe.Pointer) float64 `ffi:"lua_tonumberx,gte=503"`
	LuaTointegerx  func(L unsafe.Pointer, idx int, isnum unsafe.Pointer) int64   `ffi:"lua_tointegerx,gte=503"`
	LuaTolstring   func(L unsafe.Pointer, idx int, sz unsafe.Pointer) *byte      `ffi:"lua_tolstring,gte=503"`
	LuaToboolean   func(L unsafe.Pointer, idx int) int                           `ffi:"lua_toboolean,gte=503"`
	LuaRawlen      func(L unsafe.Pointer, idx int) uint                          `ffi:"lua_rawlen,gte=503"`
	LuaTocfunction func(L unsafe.Pointer, idx int) unsafe.Pointer                `ffi:"lua_tocfunction,gte=503"`
	LuaTouserdata  func(L unsafe.Pointer, idx int) unsafe.Pointer                `ffi:"lua_touserdata,gte=503"`
	LuaTothread    func(L unsafe.Pointer, idx int) unsafe.Pointer                `ffi:"lua_tothread,gte=503"`
	LuaTopointer   func(L unsafe.Pointer, idx int) unsafe.Pointer                `ffi:"lua_topointer,gte=503"`

	LuaRawequal func(L unsafe.Pointer, idx1 int, idx2 int) int         `ffi:"lua_rawequal,gte=503"`
	LuaCompare  func(L unsafe.Pointer, idx1 int, idx2 int, op int) int `ffi:"lua_compare,gte=503"`
	LuaArith    func(L unsafe.Pointer, op int)                         `ffi:"lua_arith,gte=503"`
	LuaConcat   func(L unsafe.Pointer, n int)                          `ffi:"lua_concat,gte=503"`
	LuaLen      func(L unsafe.Pointer, idx int)                        `ffi:"lua_len,gte=503"`

	// Push functions
	LuaPushnil           func(L unsafe.Pointer)                         `ffi:"lua_pushnil,gte=503"`
	LuaPushnumber        func(L unsafe.Pointer, n float64)              `ffi:"lua_pushnumber,gte=503"`
	LuaPushinteger       func(L unsafe.Pointer, n int64)                `ffi:"lua_pushinteger,gte=503"`
	LuaPushlstring       func(L unsafe.Pointer, s *byte, len int) *byte `ffi:"lua_pushlstring,gte=503"`
	LuaPushstring        func(L unsafe.Pointer, s *byte) *byte          `ffi:"lua_pushstring,gte=503"`
	LuaPushcclousure     func(L unsafe.Pointer, f uintptr, n int)       `ffi:"lua_pushcclosure,gte=503"`
	LuaPushboolean       func(L unsafe.Pointer, b int) int              `ffi:"lua_pushboolean,gte=503"`
	LuaPushlightuserdata func(L unsafe.Pointer, p unsafe.Pointer)       `ffi:"lua_pushlightuserdata,gte=503"`
	LuaPushthread        func(L unsafe.Pointer) int                     `ffi:"lua_pushthread,gte=503"`

	// Table and field functions
	LuaCreatetable func(L unsafe.Pointer, narr, nrec int)         `ffi:"lua_createtable,gte=503"`
	LuaGettable    func(L unsafe.Pointer, idx int) int            `ffi:"lua_gettable,gte=503"`
	LuaSettable    func(L unsafe.Pointer, idx int)                `ffi:"lua_settable,gte=503"`
	LuaGetfield    func(L unsafe.Pointer, idx int, k *byte) int32 `ffi:"lua_getfield,gte=503"`
	LuaSetfield    func(L unsafe.Pointer, idx int, k *byte)       `ffi:"lua_setfield,gte=503"`
	LuaGeti        func(L unsafe.Pointer, idx int, n int64) int   `ffi:"lua_geti,gte=503"`
	LuaSeti        func(L unsafe.Pointer, idx int, n int64)       `ffi:"lua_seti,gte=503"`
	// Table raw functions
	LuaRawget  func(L unsafe.Pointer, idx int) int32                   `ffi:"lua_rawget,gte=503"`
	LuaRawset  func(L unsafe.Pointer, idx int)                         `ffi:"lua_rawset,gte=503"`
	LuaRawgeti func(L unsafe.Pointer, idx int, n int64) int32          `ffi:"lua_rawgeti,gte=503"`
	LuaRawseti func(L unsafe.Pointer, idx int, n int64)                `ffi:"lua_rawseti,gte=503"`
	LuaRawgetp func(L unsafe.Pointer, idx int, p unsafe.Pointer) int32 `ffi:"lua_rawgetp,gte=503"`
	LuaRawsetp func(L unsafe.Pointer, idx int, p unsafe.Pointer)       `ffi:"lua_rawsetp,gte=503"`
	LuaNext    func(L unsafe.Pointer, idx int) int                     `ffi:"lua_next,gte=503"`
	// Meta table functions
	LuaGetmetatable func(L unsafe.Pointer, objindex int) int `ffi:"lua_getmetatable,gte=503"`
	LuaSetmetatable func(L unsafe.Pointer, objindex int) int `ffi:"lua_setmetatable,gte=503"`

	LuaSetupvalue func(L unsafe.Pointer, idx int, n int) *byte `ffi:"lua_setupvalue,gte=503"`
	LuaGetupvalue func(L unsafe.Pointer, idx int, n int) *byte `ffi:"lua_getupvalue,gte=503"`

	// Userdata functions
	LuaNewuserdata   func(L unsafe.Pointer, sz int) unsafe.Pointer              `ffi:"lua_newuserdata,gte=503,lte=503"`
	LuaGetuservalue  func(L unsafe.Pointer, idx int) int32                      `ffi:"lua_getuservalue,gte=503,lte=503"`
	LuaSetuservalue  func(L unsafe.Pointer, idx int)                            `ffi:"lua_setuservalue,gte=503,lte=503"`
	LuaNewuserdatauv func(L unsafe.Pointer, sz int, nuvlue int) unsafe.Pointer  `ffi:"lua_newuserdatauv,gte=504"`
	LuaGetiuservalue func(L unsafe.Pointer, idx int, n int) int32               `ffi:"lua_getiuservalue,gte=504"`
	LuaSetiuservalue func(L unsafe.Pointer, idx int, n int)                     `ffi:"lua_setiuservalue,gte=504"`
	LuaLCheckudata   func(L unsafe.Pointer, ud int, tname *byte) unsafe.Pointer `ffi:"luaL_checkudata,gte=503"`
	LuaLTestudata    func(L unsafe.Pointer, ud int, tname *byte) unsafe.Pointer `ffi:"luaL_testudata,gte=503"`

	LuaGetglobal func(L unsafe.Pointer, name *byte) int32                                                   `ffi:"lua_getglobal,gte=503"`
	LuaSetglobal func(L unsafe.Pointer, name *byte)                                                         `ffi:"lua_setglobal,gte=503"`
	LuaCallk     func(L unsafe.Pointer, nargs, nresults int, ctx unsafe.Pointer, k uintptr)                 `ffi:"lua_callk,gte=503"`
	LuaPcallk    func(L unsafe.Pointer, nargs, nresults, errfunc int, ctx unsafe.Pointer, k uintptr) int    `ffi:"lua_pcallk,gte=503"`
	LuaLoad      func(L unsafe.Pointer, reader uintptr, dt unsafe.Pointer, chunkname *byte, mode *byte) int `ffi:"lua_load,gte=503"`

	LuaSetwarnf func(L unsafe.Pointer, warnf uintptr, ud unsafe.Pointer) `ffi:"lua_setwarnf,gte=504"`

	// Coroutine functions
	LuaYieldk      func(L unsafe.Pointer, nresults int, ctx unsafe.Pointer, k uintptr) int        `ffi:"lua_yieldk,gte=503"`
	LuaResume      func(L unsafe.Pointer, from unsafe.Pointer, narg int, nres unsafe.Pointer) int `ffi:"lua_resume,gte=504"`
	LuaResume503   func(L unsafe.Pointer, from unsafe.Pointer, narg int) int                      `ffi:"lua_resume,gte=503,lte=503"`
	LuaStatus      func(L unsafe.Pointer) int                                                     `ffi:"lua_status,gte=503"`
	LuaIsyieldable func(L unsafe.Pointer) int                                                     `ffi:"lua_isyieldable,gte=503"`

	LuaLNewstate func() unsafe.Pointer `ffi:"luaL_newstate"`
	// Open all preloaded libraries.
	LuaLOpenlibs func(L unsafe.Pointer) `ffi:"luaL_openlibs"`

	LuaLNewmetatable func(L unsafe.Pointer, tname *byte) int        `ffi:"luaL_newmetatable,gte=503"`
	LuaLSetmetatable func(L unsafe.Pointer, tname *byte)            `ffi:"luaL_setmetatable,gte=503"`
	LuaLCallmeta     func(L unsafe.Pointer, ojbj int, e *byte) int  `ffi:"luaL_callmeta,gte=503"`
	LuaLGetmetafield func(L unsafe.Pointer, obj int, e *byte) int32 `ffi:"luaL_getmetafield,gte=503"`

	// Auxiliary functions
	LuaLChecknumber  func(L unsafe.Pointer, idx int) float64                             `ffi:"luaL_checknumber,gte=503"`
	LuaLCheckinteger func(L unsafe.Pointer, idx int) int64                               `ffi:"luaL_checkinteger,gte=503"`
	LuaLChecklstring func(L unsafe.Pointer, idx int, sz unsafe.Pointer) *byte            `ffi:"luaL_checklstring,gte=503"`
	LuaLChecktype    func(L unsafe.Pointer, idx int, t int)                              `ffi:"luaL_checktype,gte=503"`
	LuaLCheckany     func(L unsafe.Pointer, idx int)                                     `ffi:"luaL_checkany,gte=503"`
	LuaLOptnumber    func(L unsafe.Pointer, idx int, def float64) float64                `ffi:"luaL_optnumber,gte=503"`
	LuaLOptinteger   func(L unsafe.Pointer, idx int, def int64) int64                    `ffi:"luaL_optinteger,gte=503"`
	LuaLOptlstring   func(L unsafe.Pointer, idx int, def *byte, sz unsafe.Pointer) *byte `ffi:"luaL_optlstring,gte=503"`
	LuaLCheckstack   func(L unsafe.Pointer, sz int, msg *byte) int                       `ffi:"luaL_checkstack,gte=503"`
	LuaLTolstring    func(L unsafe.Pointer, idx int, sz unsafe.Pointer) *byte            `ffi:"luaL_tolstring,gte=503"`

	LuaLError       func(L unsafe.Pointer, msg *byte) int                                  `ffi:"luaL_error,gte=503"`
	LuaLLoadstring  func(L unsafe.Pointer, s *byte) int                                    `ffi:"luaL_loadstring,gte=503"`
	LuaLLoadfilex   func(L unsafe.Pointer, filename *byte, mode *byte) int                 `ffi:"luaL_loadfilex,gte=503"`
	LuaLLoadbufferx func(L unsafe.Pointer, buff *byte, sz int, name *byte, mode *byte) int `ffi:"luaL_loadbufferx,gte=503"`

	LuaLSetfuncs func(L unsafe.Pointer, l unsafe.Pointer, nup int) `ffi:"luaL_setfuncs,gte=503"`

	LuaLTraceback func(L unsafe.Pointer, L1 unsafe.Pointer, msg *byte, level int) int `ffi:"luaL_traceback,gte=503"`

	LuaLRef      func(L unsafe.Pointer, idx int) int                           `ffi:"luaL_ref,gte=503"`
	LuaLUnref    func(L unsafe.Pointer, idx int, ref int)                      `ffi:"luaL_unref,gte=503"`
	LuaLRequiref func(L unsafe.Pointer, modname *byte, openf uintptr, glb int) `ffi:"luaL_requiref,gte=503"`
}

// Lib returns the underlying dynamic library handle for this ffi instance.
func (ffi *ffi) Lib() uintptr {
	return ffi.lib
}

// getLuaVersion retrieves the Lua version from the loaded library.
// It uses the lua_version function to determine the version number.
// With this, we can conditionally register functions based on the Lua version.
func getLuaVersion(lib uintptr) (version float64) {
	var (
		luaVersion   func(L unsafe.Pointer) float64
		luaLNewState func() unsafe.Pointer
		luaClose     func(L unsafe.Pointer)
	)
	purego.RegisterLibFunc(&luaVersion, lib, "lua_version")
	purego.RegisterLibFunc(&luaLNewState, lib, "luaL_newstate")
	purego.RegisterLibFunc(&luaClose, lib, "lua_close")
	L := luaLNewState()
	defer luaClose(L)

	version = luaVersion(L)
	// Lower than Lua 5.4
	// In lua5.3 windows, the previous version is returned as 1.010407847501e-311
	// In other platforms it returns as 0
	if version < 0.1 {
		var luaVersion func(L unsafe.Pointer) *float64
		purego.RegisterLibFunc(&luaVersion, lib, "lua_version")
		version = *luaVersion(L)
	}
	return
}

// newFFI loads the Lua 5.4 dynamic library at the specified path and registers all available exported entrypoints.
// It provides a Go ffi struct ready for low-level Lua C API interaction in memory.
func newFFI(path string) (FFI *ffi, err error) {
	lib, err := loadLibrary(path)
	if err != nil {
		return
	}

	ver := getLuaVersion(lib)

	FFI = &ffi{
		lib:     lib,
		version: ver,
	}

	version := int(ver)

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
				}
			case "lte":
				if version > targetVersion {
					register = false
				}
			}
			if !register {
				break
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
