package lua_test

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/lua/lua54"
)

type testObject struct {
	data [1024]byte
	id   int
}

func (s *Suite) TestTypeLightUserData(assert *require.Assertions, L *lua.State) {

	var testVar = &testObject{id: 123}
	testVar.data[0] = 42
	err := L.PushLightUserData(testVar)
	assert.NoError(err)

	assert.True(L.IsLightUserData(-1))
	assert.Equal(lua.LUA_TLIGHTUSERDATA, L.Type(-1))
	assert.Equal("userdata", L.TypeName(L.Type(-1)))

	assert.True(L.IsUserData(-1))

	ptr := L.ToUserData(-1)
	assert.NotNil(ptr)

	retrievedPtr := (*testObject)(ptr)
	assert.Equal(byte(42), retrievedPtr.data[0])
	assert.Equal(123, retrievedPtr.id)

	L.Pop(1)
	unsafePtr := unsafe.Pointer(&testVar)
	err = L.PushLightUserData(unsafePtr)
	assert.NoError(err)

	assert.True(L.IsLightUserData(-1))
	retrievedUnsafePtr := L.ToUserData(-1)
	assert.Equal(unsafePtr, retrievedUnsafePtr)

	L.Pop(1)
	err = L.PushLightUserData(42)
	assert.Error(err)

	var nilPtr *int
	err = L.PushLightUserData(nilPtr)
	assert.NoError(err)
	assert.True(L.IsLightUserData(-1))

	nilPtrRetrieved := L.ToUserData(-1)
	assert.Nil(nilPtrRetrieved)
}

func (s *Suite) TestTypeToCFunction(assert *require.Assertions, L *lua.State) {

	testCFunc := func(L *lua.State) int {
		L.PushString("Hello from C function")
		return 1
	}

	L.PushCFunction(testCFunc)

	assert.True(L.IsCFunction(-1))
	assert.Equal(lua.LUA_TFUNCTION, L.Type(-1))
	assert.Equal("function", L.TypeName(L.Type(-1)))

	cfuncPtr := L.ToCFunction(-1)
	assert.NotNil(cfuncPtr)

	L.Pop(1)
	err := L.LoadString("function test() return 42 end")
	assert.NoError(err)

	assert.True(L.IsFunction(-1))
	assert.False(L.IsCFunction(-1))

	luaFuncPtr := L.ToCFunction(-1)
	assert.Nil(luaFuncPtr)

	L.Pop(1)
	L.PushInteger(42)
	assert.False(L.IsCFunction(-1))
	nonFuncPtr := L.ToCFunction(-1)
	assert.Nil(nonFuncPtr)

	L.Pop(1)
	L.PushNil()
	assert.False(L.IsCFunction(-1))
	nilPtr := L.ToCFunction(-1)
	assert.Nil(nilPtr)
}

func (s *Suite) TestTypeToRawLen(assert *require.Assertions, L *lua.State) {

	testString := "Hello, World!"
	L.PushString(testString)
	rawLen := L.RawLen(-1)
	assert.Equal(uint(len(testString)), rawLen)

	L.Pop(1)

	L.PushString("")
	rawLen = L.RawLen(-1)
	assert.Equal(uint(0), rawLen)

	L.Pop(1)

	longString := "This is a longer test string with more characters"
	L.PushString(longString)
	rawLen = L.RawLen(-1)
	assert.Equal(uint(len(longString)), rawLen)

	L.Pop(1)

	L.PushInteger(42)

	rawLen = L.RawLen(-1)
	assert.Equal(uint(0), rawLen)

	L.Pop(1)

	L.PushBoolean(true)
	rawLen = L.RawLen(-1)
	assert.Equal(uint(0), rawLen)

	L.Pop(1)

	L.PushNil()
	rawLen = L.RawLen(-1)
	assert.Equal(uint(0), rawLen)

	L.Pop(1)

	var testVar int = 42
	err := L.PushLightUserData(&testVar)
	assert.NoError(err)

	rawLen = L.RawLen(-1)
	assert.Equal(uint(0), rawLen)
}

func (s *Suite) TestFunction(assert *require.Assertions, L *lua.State) {
	assert.Equal(fmt.Sprintf("%.0f", L.Version()), "504")

	L.PushCFunction(func(L *lua.State) int {
		number := L.ToNumber(1)
		assert.Equal(number, 42.0)
		return 0
	})
	assert.NoError(L.SetGlobal("print_number"))

	L.PushCFunction(func(L *lua.State) int {
		x := L.CheckNumber(1)
		L.PushNumber(x * 2)
		return 1
	})
	assert.NoError(L.SetGlobal("double_number"))

	assert.NoError(L.DoString(`print_number(double_number(21))`))
}

func (s *Suite) TestCheckNumber(assert *require.Assertions, L *lua.State) {

	L.PushNumber(42.5)
	result := L.CheckNumber(-1)
	assert.Equal(42.5, result)
	L.Pop(1)

	L.PushInteger(123)
	result = L.CheckNumber(-1)
	assert.Equal(123.0, result)
	L.Pop(1)

	L.PushString("456.7")
	result = L.CheckNumber(-1)
	assert.Equal(456.7, result)
	L.Pop(1)

	L.PushNumber(0.0)
	result = L.CheckNumber(-1)
	assert.Equal(0.0, result)
	L.Pop(1)

	L.PushNumber(-123.45)
	result = L.CheckNumber(-1)
	assert.Equal(-123.45, result)
	L.Pop(1)

	L.AtPanic(func(L *lua.State) int {
		err := L.CheckError(lua.LUA_ERRERR)
		assert.Error(err)

		panic(err)
	})
	L.PushString("not a number")
	assert.Panics(func() {
		L.CheckNumber(-1)
	})
}

func (s *Suite) TestCheckInteger(assert *require.Assertions, L *lua.State) {

	L.PushInteger(42)
	result := L.CheckInteger(-1)
	assert.Equal(int64(42), result)
	L.Pop(1)

	L.PushNumber(123.0)
	result = L.CheckInteger(-1)
	assert.Equal(int64(123), result)
	L.Pop(1)

	L.PushString("456")
	result = L.CheckInteger(-1)
	assert.Equal(int64(456), result)
	L.Pop(1)

	L.PushInteger(0)
	result = L.CheckInteger(-1)
	assert.Equal(int64(0), result)
	L.Pop(1)

	L.PushInteger(-789)
	result = L.CheckInteger(-1)
	assert.Equal(int64(-789), result)
	L.Pop(1)

	L.PushInteger(math.MaxInt64)
	result = L.CheckInteger(-1)
	assert.Equal(int64(math.MaxInt64), result)
	L.Pop(1)
}

func (s *Suite) TestCheckLString(assert *require.Assertions, L *lua.State) {

	testStr := "Hello, World!"
	L.PushString(testStr)
	size := len(testStr)
	result := L.CheckLString(-1, size)
	assert.Equal(testStr, result)
	L.Pop(1)

	L.PushString("")
	result = L.CheckLString(-1, 0)
	assert.Equal("", result)
	L.Pop(1)

	unicodeStr := "你好世界🌍"
	L.PushString(unicodeStr)
	size = len(unicodeStr)
	result = L.CheckLString(-1, size)
	assert.Equal(unicodeStr, result)
	L.Pop(1)

	L.PushNumber(42.5)
	result = L.CheckLString(-1, 10)
	assert.Equal("42.5", result)
	L.Pop(1)

	L.PushInteger(123)
	result = L.CheckLString(-1, 10)
	assert.Equal("123", result)
	L.Pop(1)
}

func (s *Suite) TestCheckType(assert *require.Assertions, L *lua.State) {

	L.PushNumber(42.5)

	L.CheckType(-1, lua.LUA_TNUMBER)
	L.Pop(1)

	L.PushString("test")
	L.CheckType(-1, lua.LUA_TSTRING)
	L.Pop(1)

	L.PushBoolean(true)
	L.CheckType(-1, lua.LUA_TBOOLEAN)
	L.Pop(1)

	L.PushNil()
	L.CheckType(-1, lua.LUA_TNIL)
	L.Pop(1)

	L.PushCFunction(func(L *lua.State) int { return 0 })
	L.CheckType(-1, lua.LUA_TFUNCTION)
	L.Pop(1)

	L.AtPanic(func(L *lua.State) int {
		err := L.CheckError(lua.LUA_ERRERR)
		assert.Error(err)

		panic(err)
	})

	L.PushNumber(42.5)
	assert.Panics(func() {
		L.CheckType(-1, lua.LUA_TSTRING)
	})
}

func (s *Suite) TestCheckAny(assert *require.Assertions, L *lua.State) {

	L.PushNumber(42.5)

	L.CheckAny(-1)
	L.Pop(1)

	L.PushString("test")
	L.CheckAny(-1)
	L.Pop(1)

	L.PushBoolean(false)
	L.CheckAny(-1)
	L.Pop(1)

	L.PushNil()
	L.CheckAny(-1)
	L.Pop(1)

	L.PushCFunction(func(L *lua.State) int { return 0 })
	L.CheckAny(-1)
	L.Pop(1)

	L.AtPanic(func(L *lua.State) int {
		err := L.CheckError(lua.LUA_ERRERR)
		assert.Error(err)

		panic(err)
	})

	assert.Panics(func() {
		L.CheckAny(1)
	})
}

func (s *Suite) TestOptNumber(assert *require.Assertions, L *lua.State) {

	L.PushNumber(42.5)
	result := L.OptNumber(-1, 100.0)
	assert.Equal(42.5, result)
	L.Pop(1)

	L.PushInteger(123)
	result = L.OptNumber(-1, 100.0)
	assert.Equal(123.0, result)
	L.Pop(1)

	L.PushString("456.7")
	result = L.OptNumber(-1, 100.0)
	assert.Equal(456.7, result)
	L.Pop(1)

	L.PushNil()
	result = L.OptNumber(-1, 999.99)
	assert.Equal(999.99, result)
	L.Pop(1)

	L.PushNumber(0.0)
	result = L.OptNumber(-1, 100.0)
	assert.Equal(0.0, result)
	L.Pop(1)

	L.PushNumber(-123.45)
	result = L.OptNumber(-1, 100.0)
	assert.Equal(-123.45, result)
	L.Pop(1)

	result = L.OptNumber(100, 777.77)
	assert.Equal(777.77, result)

	L.PushNil()
	result = L.OptNumber(-1, -500.5)
	assert.Equal(-500.5, result)
	L.Pop(1)
}

func (s *Suite) TestOptInteger(assert *require.Assertions, L *lua.State) {

	L.PushInteger(42)
	result := L.OptInteger(-1, 100)
	assert.Equal(int64(42), result)
	L.Pop(1)

	L.PushNumber(123.0)
	result = L.OptInteger(-1, 100)
	assert.Equal(int64(123), result)
	L.Pop(1)

	L.PushString("456")
	result = L.OptInteger(-1, 100)
	assert.Equal(int64(456), result)
	L.Pop(1)

	L.PushNil()
	result = L.OptInteger(-1, 999)
	assert.Equal(int64(999), result)
	L.Pop(1)

	L.PushInteger(0)
	result = L.OptInteger(-1, 100)
	assert.Equal(int64(0), result)
	L.Pop(1)

	L.PushInteger(-789)
	result = L.OptInteger(-1, 100)
	assert.Equal(int64(-789), result)
	L.Pop(1)

	L.PushInteger(math.MaxInt64)
	result = L.OptInteger(-1, 100)
	assert.Equal(int64(math.MaxInt64), result)
	L.Pop(1)

	result = L.OptInteger(100, 777)
	assert.Equal(int64(777), result)

	L.PushNil()
	result = L.OptInteger(-1, -500)
	assert.Equal(int64(-500), result)
	L.Pop(1)
}

func (s *Suite) TestOptLString(assert *require.Assertions, L *lua.State) {

	testStr := "Hello, World!"
	L.PushString(testStr)
	var size int
	result, err := L.OptLString(-1, "default", &size)
	assert.NoError(err)
	assert.Equal(testStr, result)
	assert.Equal(len(testStr), size)
	L.Pop(1)

	L.PushString("")
	result, err = L.OptLString(-1, "default", &size)
	assert.NoError(err)
	assert.Equal("", result)
	assert.Equal(0, size)
	L.Pop(1)

	unicodeStr := "你好世界🌍"
	L.PushString(unicodeStr)
	result, err = L.OptLString(-1, "default", &size)
	assert.NoError(err)
	assert.Equal(unicodeStr, result)
	assert.Equal(len(unicodeStr), size)
	L.Pop(1)

	L.PushNumber(42.5)
	result, err = L.OptLString(-1, "default", &size)
	assert.NoError(err)
	assert.Equal("42.5", result)
	assert.True(size > 0)
	L.Pop(1)

	L.PushInteger(123)
	result, err = L.OptLString(-1, "default", &size)
	assert.NoError(err)
	assert.Equal("123", result)
	assert.Equal(3, size)
	L.Pop(1)

	L.PushNil()
	defaultStr := "this is default"
	result, err = L.OptLString(-1, defaultStr, &size)
	assert.NoError(err)
	assert.Equal(defaultStr, result)
	assert.Equal(len(defaultStr), size)
	L.Pop(1)

	defaultStr2 := "another default"
	result, err = L.OptLString(100, defaultStr2, &size)
	assert.NoError(err)
	assert.Equal(defaultStr2, result)
	assert.Equal(len(defaultStr2), size)

	L.PushString("test")
	result, err = L.OptLString(-1, "default", nil)
	assert.NoError(err)
	assert.Equal("test", result)
	L.Pop(1)

	L.PushNil()
	result, err = L.OptLString(-1, "", &size)
	assert.NoError(err)
	assert.Equal("", result)
	assert.Equal(0, size)
	L.Pop(1)

	longStr := "This is a very long string that contains many characters and should test the string handling properly in the Lua to Go conversion"
	L.PushString(longStr)
	result, err = L.OptLString(-1, "default", &size)
	assert.NoError(err)
	assert.Equal(longStr, result)
	assert.Equal(len(longStr), size)
	L.Pop(1)
}
