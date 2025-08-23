package lua_test

import (
	"math"
	"unsafe"

	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/lua"
)

type testObject struct {
	data [1024]byte
	id   int
}

func (s *Suite) TestTypeLightUserData(assert *require.Assertions, L *lua.State) {
	var testVar = &testObject{id: 123}
	testVar.data[0] = 42
	L.PushLightUserData(testVar)

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
	L.PushLightUserData(unsafePtr)

	assert.True(L.IsLightUserData(-1))
	retrievedUnsafePtr := L.ToUserData(-1)
	assert.Equal(unsafePtr, retrievedUnsafePtr)

	L.Pop(1)
	assert.Panics(func() {
		L.PushLightUserData(42)
	})

	var nilPtr *int
	L.PushLightUserData(nilPtr)
	assert.True(L.IsLightUserData(-1))

	nilPtrRetrieved := L.ToUserData(-1)
	assert.Nil(nilPtrRetrieved)
}

func (s *Suite) TestTypeToPointer(assert *require.Assertions, L *lua.State) {
	var testVar = &testObject{id: 456}
	testVar.data[0] = 84
	L.PushLightUserData(testVar)

	assert.True(L.IsLightUserData(-1))
	assert.Equal(lua.LUA_TLIGHTUSERDATA, L.Type(-1))
	assert.Equal("userdata", L.TypeName(L.Type(-1)))

	ptr := L.ToPointer(-1)
	assert.NotNil(ptr)

	retrievedPtr := (*testObject)(ptr)
	assert.Equal(byte(84), retrievedPtr.data[0])
	assert.Equal(456, retrievedPtr.id)

	L.Pop(1)

	// Test with nil pointer
	var nilPtr *int
	L.PushLightUserData(nilPtr)
	nilPtrRetrieved := L.ToPointer(-1)
	assert.Nil(nilPtrRetrieved)

	L.Pop(1)
}

func (s *Suite) TestTypeToGoFunction(assert *require.Assertions, L *lua.State) {
	expected := "Hello from Go function"
	testCFunc := func(L *lua.State) int {
		L.PushString(expected)
		return 1
	}

	L.PushGoFunction(testCFunc)

	assert.True(L.IsGoFunction(-1))
	assert.Equal(lua.LUA_TFUNCTION, L.Type(-1))
	assert.Equal("function", L.TypeName(L.Type(-1)))

	cFunc := L.ToCFunction(-1)
	assert.NotNil(cFunc)

	assert.NoError(L.PCall(0, 1, 0))
	assert.Equal(expected, L.ToString(-1))
	L.Pop(1)

	L.PushCFunction(cFunc)
	assert.NoError(L.PCall(0, 1, 0))
	assert.Equal(expected, L.ToString(-1))
	L.Pop(1)

	err := L.LoadString("function test() return 42 end")
	assert.NoError(err)

	assert.True(L.IsFunction(-1))
	assert.False(L.IsGoFunction(-1))

	luaFunc := L.ToCFunction(-1)
	assert.Nil(luaFunc)

	L.Pop(1)
	L.PushInteger(42)
	assert.False(L.IsGoFunction(-1))
	nonFunc := L.ToCFunction(-1)
	assert.Nil(nonFunc)

	L.Pop(1)
	L.PushNil()
	assert.False(L.IsGoFunction(-1))
	nilFunc := L.ToCFunction(-1)
	assert.Nil(nilFunc)
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

	var testVar = 42
	L.PushLightUserData(&testVar)

	rawLen = L.RawLen(-1)
	assert.Equal(uint(0), rawLen)
}

func (s *Suite) TestFunction(assert *require.Assertions, L *lua.State) {
	L.PushGoFunction(func(L *lua.State) int {
		number := L.ToNumber(1)
		assert.Equal(number, 42.0)
		return 0
	})
	L.SetGlobal("print_number")

	L.PushGoFunction(func(L *lua.State) int {
		x := L.CheckNumber(1)
		L.PushNumber(x * 2)
		return 1
	})
	L.SetGlobal("double_number")

	assert.NoError(L.DoString(`print_number(double_number(21))`))
}

func (s *Suite) TestFunctionUpValue(assert *require.Assertions, L *lua.State) {
	L.PushString("World")
	L.PushGoClousure(func(L *lua.State) int {
		upValue := L.ToString(L.UpValueIndex(1))
		L.PushString("Hello, " + upValue)
		return 1
	}, 1)
	L.PushValue(-1)
	assert.NoError(L.PCall(0, 1, 0))
	assert.Equal("Hello, World", L.ToString(-1))
	L.Pop(1)

	L.PushString("Go")
	L.SetUpValue(-2, 1)
	assert.Empty(L.GetUpValue(-1, 1))
	assert.Equal("Go", L.ToString(-1))
	L.Pop(1)

	assert.NoError(L.PCall(0, 1, 0))
	assert.Equal("Hello, Go", L.ToString(-1))

	err := L.DoString(`
function make_counter()
  local count = 0
	return function()
	  count = count + 1
		return count
	end
end
counter = make_counter()
	`)
	assert.NoError(err)
	L.GetGlobal("counter")
	assert.Equal("count", L.GetUpValue(-1, 1))
	L.Pop(1)
	L.PushInteger(2)
	assert.Equal("count", L.SetUpValue(-2, 1))
	assert.NoError(L.PCall(0, 1, 0))
	assert.EqualValues(3, L.ToInteger(-1))
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
	var size int
	result := L.CheckLString(-1, &size)
	assert.Equal(testStr, result)
	assert.Equal(len(testStr), size)
	L.Pop(1)

	L.PushString(testStr)
	result = L.CheckString(-1)
	assert.Equal(testStr, result)
	L.Pop(1)

	L.PushString("")
	result = L.CheckLString(-1, &size)
	assert.Equal("", result)
	assert.Equal(0, size)
	L.Pop(1)

	unicodeStr := "你好世界🌍"
	L.PushString(unicodeStr)
	size = len(unicodeStr)
	result = L.CheckLString(-1, &size)
	assert.Equal(unicodeStr, result)
	assert.Equal(len(unicodeStr), size)
	L.Pop(1)

	L.PushNumber(42.5)
	result = L.CheckLString(-1, &size)
	assert.Equal("42.5", result)
	assert.Equal(len("42.5"), size)
	L.Pop(1)

	L.PushInteger(123)
	result = L.CheckLString(-1, &size)
	assert.Equal("123", result)
	assert.Equal(len("123"), size)
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

	L.PushGoFunction(func(L *lua.State) int { return 0 })
	L.CheckType(-1, lua.LUA_TFUNCTION)
	L.Pop(1)

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

	L.PushGoFunction(func(L *lua.State) int { return 0 })
	L.CheckAny(-1)
	L.Pop(1)

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
	result := L.OptLString(-1, "default", &size)
	assert.Equal(testStr, result)
	assert.Equal(len(testStr), size)
	L.Pop(1)

	L.PushString("")
	result = L.OptLString(-1, "default", &size)
	assert.Equal("", result)
	assert.Equal(0, size)
	L.Pop(1)

	unicodeStr := "你好世界🌍"
	L.PushString(unicodeStr)
	result = L.OptLString(-1, "default", &size)
	assert.Equal(unicodeStr, result)
	assert.Equal(len(unicodeStr), size)
	L.Pop(1)

	L.PushNumber(42.5)
	result = L.OptLString(-1, "default", &size)
	assert.Equal("42.5", result)
	assert.True(size > 0)
	L.Pop(1)

	L.PushInteger(123)
	result = L.OptLString(-1, "default", &size)
	assert.Equal("123", result)
	assert.Equal(3, size)
	L.Pop(1)

	L.PushNil()
	defaultStr := "this is default"
	result = L.OptLString(-1, defaultStr, &size)
	assert.Equal(defaultStr, result)
	assert.Equal(len(defaultStr), size)
	L.Pop(1)

	defaultStr2 := "another default"
	result = L.OptLString(100, defaultStr2, &size)
	assert.Equal(defaultStr2, result)
	assert.Equal(len(defaultStr2), size)

	L.PushString("test")
	result = L.OptLString(-1, "default", nil)
	assert.Equal("test", result)
	L.Pop(1)

	L.PushNil()
	result = L.OptLString(-1, "", &size)
	assert.Equal("", result)
	assert.Equal(0, size)
	L.Pop(1)

	longStr := "This is a very long string that contains many characters and should test the string handling properly in the Lua to Go conversion"
	L.PushString(longStr)
	result = L.OptLString(-1, "default", &size)
	assert.Equal(longStr, result)
	assert.Equal(len(longStr), size)
	L.Pop(1)
}
