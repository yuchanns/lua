package lua

import (
	"reflect"
	"unsafe"

	"github.com/ebitengine/purego"
	"go.yuchanns.xyz/lua/internal/tools"
)

type LuaAlloc func(ud unsafe.Pointer, ptr unsafe.Pointer, osize, nsize int) unsafe.Pointer

type LuaCFunction func(L unsafe.Pointer) int

type LuaKFunction func(L unsafe.Pointer, status int, ctx int) int

type LuaWarnFunction func(ud unsafe.Pointer, msg *byte, tocont int)

type ffi struct {
	lib uintptr

	// State manipulation
	LuaNewstate func(f LuaAlloc, ud unsafe.Pointer) unsafe.Pointer `ffi:"lua_newstate"`
	LuaClose    func(L unsafe.Pointer)                             `ffi:"lua_close"`

	LuaAtpanic func(L unsafe.Pointer, panicf LuaCFunction) unsafe.Pointer `ffi:"lua_atpanic"`

	LuaVersion func(L unsafe.Pointer) float64 `ffi:"lua_version"`

	LuaPushcclousure func(L unsafe.Pointer, f LuaCFunction, n int)                                     `ffi:"lua_pushcclosure"`
	LuaSetglobal     func(L unsafe.Pointer, name *byte)                                                `ffi:"lua_setglobal"`
	LuaGettop        func(L unsafe.Pointer) int                                                        `ffi:"lua_gettop"`
	LuaSettop        func(L unsafe.Pointer, idx int)                                                   `ffi:"lua_settop"`
	LuaPushnumber    func(L unsafe.Pointer, n float64) int                                             `ffi:"lua_pushnumber"`
	LuaPushinteger   func(L unsafe.Pointer, n int64) int                                               `ffi:"lua_pushinteger"`
	LuaPushlstring   func(L unsafe.Pointer, s *byte, len int) int                                      `ffi:"lua_pushlstring"`
	LuaPushboolean   func(L unsafe.Pointer, b int) int                                                 `ffi:"lua_pushboolean"`
	LuaPcallk        func(L unsafe.Pointer, nargs, nresults, errfunc int, ctx int, k LuaKFunction) int `ffi:"lua_pcallk"`
	LuaTolstring     func(L unsafe.Pointer, idx int, size unsafe.Pointer) *byte                        `ffi:"lua_tolstring"`
	LuaToboolean     func(L unsafe.Pointer, idx int) int                                               `ffi:"lua_toboolean"`

	LuaSetwarnf func(L unsafe.Pointer, warnf LuaWarnFunction, ud unsafe.Pointer) `ffi:"lua_setwarnf"`

	LuaLNewstate func() unsafe.Pointer `ffi:"luaL_newstate"`
	// Open all preloaded libraries.
	LuaLOpenlibs     func(L unsafe.Pointer)                                    `ffi:"luaL_openlibs"`
	LuaLChecknumber  func(L unsafe.Pointer, idx int) float64                   `ffi:"luaL_checknumber"`
	LuaLCheckinteger func(L unsafe.Pointer, idx int) int64                     `ffi:"luaL_checkinteger"`
	LuaLChecklstring func(L unsafe.Pointer, idx int, len unsafe.Pointer) *byte `ffi:"luaL_checklstring"`
	LuaLLoadstring   func(L unsafe.Pointer, s *byte) int                       `ffi:"luaL_loadstring"`
	LuaLChecktype    func(L unsafe.Pointer, idx int, t int)                    `ffi:"luaL_checktype"`
	LuaLError        func(L unsafe.Pointer, msg *byte) int                     `ffi:"luaL_error"`
}

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
