package lua_test

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/lua/lua54"
)

func (s *Suite) TestTypeLightUserData(assert *require.Assertions, L *lua.State) {
	// Test pushing light user data
	var testVar int = 42
	err := L.PushLightUserData(&testVar)
	assert.NoError(err)

	// Verify the type is correct
	assert.True(L.IsLightUserData(-1))
	assert.Equal(lua.LUA_TLIGHTUSERDATA, L.Type(-1))
	assert.Equal("userdata", L.TypeName(L.Type(-1)))

	// Verify IsUserData also returns true (light userdata is a kind of userdata)
	assert.True(L.IsUserData(-1))

	// Get the pointer back and verify it's the same
	ptr := L.ToUserData(-1)
	assert.NotNil(ptr)

	// Convert back to the original pointer and verify the value
	retrievedPtr := (*int)(ptr)
	assert.Equal(42, *retrievedPtr)
	assert.Equal(&testVar, retrievedPtr)

	// Test with unsafe.Pointer directly
	L.Pop(1) // Remove the previous value
	unsafePtr := unsafe.Pointer(&testVar)
	err = L.PushLightUserData(unsafePtr)
	assert.NoError(err)

	assert.True(L.IsLightUserData(-1))
	retrievedUnsafePtr := L.ToUserData(-1)
	assert.Equal(unsafePtr, retrievedUnsafePtr)

	// Test error case - non-pointer value
	L.Pop(1)                      // Remove the previous value
	err = L.PushLightUserData(42) // This should fail
	assert.Error(err)

	// Test with nil pointer
	var nilPtr *int
	err = L.PushLightUserData(nilPtr)
	assert.NoError(err)
	assert.True(L.IsLightUserData(-1))

	nilPtrRetrieved := L.ToUserData(-1)
	assert.Nil(nilPtrRetrieved)
}

func (s *Suite) TestTypeToCFunction(assert *require.Assertions, L *lua.State) {
	// Test ToCFunction with a C function
	testCFunc := func(L *lua.State) int {
		L.PushString("Hello from C function")
		return 1
	}

	// Push the C function
	L.PushCFunction(testCFunc)

	// Verify the type is correct
	assert.True(L.IsCFunction(-1))
	assert.Equal(lua.LUA_TFUNCTION, L.Type(-1))
	assert.Equal("function", L.TypeName(L.Type(-1)))

	// Test ToCFunction - should return a non-nil pointer for C functions
	cfuncPtr := L.ToCFunction(-1)
	assert.NotNil(cfuncPtr)

	// Test with non-C function (Lua function)
	L.Pop(1) // Remove C function
	err := L.LoadString("function test() return 42 end")
	assert.NoError(err)

	// This creates a Lua function, not a C function
	assert.True(L.IsFunction(-1))
	assert.False(L.IsCFunction(-1))

	// ToCFunction should return nil for Lua functions
	luaFuncPtr := L.ToCFunction(-1)
	assert.Nil(luaFuncPtr)

	// Test with non-function types
	L.Pop(1) // Remove Lua function
	L.PushInteger(42)
	assert.False(L.IsCFunction(-1))
	nonFuncPtr := L.ToCFunction(-1)
	assert.Nil(nonFuncPtr)

	// Test with nil
	L.Pop(1)
	L.PushNil()
	assert.False(L.IsCFunction(-1))
	nilPtr := L.ToCFunction(-1)
	assert.Nil(nilPtr)
}

func (s *Suite) TestTypeToRawLen(assert *require.Assertions, L *lua.State) {
	// Test ToRawLen with string
	testString := "Hello, World!"
	L.PushString(testString)
	rawLen := L.ToRawLen(-1)
	assert.Equal(len(testString), rawLen)

	L.Pop(1)

	// Test with empty string
	L.PushString("")
	rawLen = L.ToRawLen(-1)
	assert.Equal(0, rawLen)

	L.Pop(1)

	// Test with longer string
	longString := "This is a longer test string with more characters"
	L.PushString(longString)
	rawLen = L.ToRawLen(-1)
	assert.Equal(len(longString), rawLen)

	L.Pop(1)

	// Test with non-string, non-table types
	L.PushInteger(42)
	// Numbers don't have meaningful raw length in Lua 5.4, should return 0
	rawLen = L.ToRawLen(-1)
	assert.Equal(0, rawLen)

	L.Pop(1)

	// Test with boolean
	L.PushBoolean(true)
	rawLen = L.ToRawLen(-1)
	assert.Equal(0, rawLen)

	L.Pop(1)

	// Test with nil
	L.PushNil()
	rawLen = L.ToRawLen(-1)
	assert.Equal(0, rawLen)

	L.Pop(1)

	// Test with userdata (if it has a __len metamethod, it might have length)
	var testVar int = 42
	err := L.PushLightUserData(&testVar)
	assert.NoError(err)
	// Light userdata typically has no raw length
	rawLen = L.ToRawLen(-1)
	assert.Equal(0, rawLen)

	// TODO: test with tables
}

func (s *Suite) TestFunction(assert *require.Assertions, L *lua.State) {
	assert.Equal(fmt.Sprintf("%.0f", L.Version()), "504")

	L.PushCFunction(func(L *lua.State) int {
		number := L.ToNumber(1)
		assert.Equal(number, 42.0)
		return 0
	})
	assert.NoError(L.SetGlobal("print_number"))

	L.PushGoFunction(func(x float64) float64 {
		return x * 2
	})
	assert.NoError(L.SetGlobal("double_number"))

	assert.NoError(L.DoString(`print_number(double_number(21))`))
}

func (s *Suite) TestCheckNumber(assert *require.Assertions, L *lua.State) {
	// Test CheckNumber with valid number
	L.PushNumber(42.5)
	result := L.CheckNumber(-1)
	assert.Equal(42.5, result)
	L.Pop(1)

	// Test CheckNumber with integer (should work as integers are numbers)
	L.PushInteger(123)
	result = L.CheckNumber(-1)
	assert.Equal(123.0, result)
	L.Pop(1)

	// Test CheckNumber with string number (should work as Lua coerces)
	L.PushString("456.7")
	result = L.CheckNumber(-1)
	assert.Equal(456.7, result)
	L.Pop(1)

	// Test CheckNumber with zero
	L.PushNumber(0.0)
	result = L.CheckNumber(-1)
	assert.Equal(0.0, result)
	L.Pop(1)

	// Test CheckNumber with negative number
	L.PushNumber(-123.45)
	result = L.CheckNumber(-1)
	assert.Equal(-123.45, result)
	L.Pop(1)

	// Note: CheckNumber should panic/error with non-numeric types
	// but we can't easily test that in this test framework as it would
	// cause the Lua state to error out and potentially terminate
}

func (s *Suite) TestCheckInteger(assert *require.Assertions, L *lua.State) {
	// Test CheckInteger with valid integer
	L.PushInteger(42)
	result := L.CheckInteger(-1)
	assert.Equal(int64(42), result)
	L.Pop(1)

	// Test CheckInteger with number that can be converted to integer
	L.PushNumber(123.0)
	result = L.CheckInteger(-1)
	assert.Equal(int64(123), result)
	L.Pop(1)

	// Test CheckInteger with string integer (should work as Lua coerces)
	L.PushString("456")
	result = L.CheckInteger(-1)
	assert.Equal(int64(456), result)
	L.Pop(1)

	// Test CheckInteger with zero
	L.PushInteger(0)
	result = L.CheckInteger(-1)
	assert.Equal(int64(0), result)
	L.Pop(1)

	// Test CheckInteger with negative integer
	L.PushInteger(-789)
	result = L.CheckInteger(-1)
	assert.Equal(int64(-789), result)
	L.Pop(1)

	// Test CheckInteger with large integer
	L.PushInteger(math.MaxInt64) // max int64
	result = L.CheckInteger(-1)
	assert.Equal(int64(math.MaxInt64), result)
	L.Pop(1)
}

func (s *Suite) TestCheckLString(assert *require.Assertions, L *lua.State) {
	// Test CheckLString with valid string
	testStr := "Hello, World!"
	L.PushString(testStr)
	size := len(testStr)
	result := L.CheckLString(-1, size)
	assert.Equal(testStr, result)
	L.Pop(1)

	// Test CheckLString with empty string
	L.PushString("")
	result = L.CheckLString(-1, 0)
	assert.Equal("", result)
	L.Pop(1)

	// Test CheckLString with Unicode string
	unicodeStr := "你好世界🌍"
	L.PushString(unicodeStr)
	size = len(unicodeStr) // byte length, not character length
	result = L.CheckLString(-1, size)
	assert.Equal(unicodeStr, result)
	L.Pop(1)

	// Test CheckLString with number (should be coerced to string)
	L.PushNumber(42.5)
	result = L.CheckLString(-1, 10) // approximate size
	assert.Equal("42.5", result)
	L.Pop(1)

	// Test CheckLString with integer (should be coerced to string)
	L.PushInteger(123)
	result = L.CheckLString(-1, 10) // approximate size
	assert.Equal("123", result)
	L.Pop(1)
}

func (s *Suite) TestCheckType(assert *require.Assertions, L *lua.State) {
	// Test CheckType with number
	L.PushNumber(42.5)
	// This should not panic/error as the type matches
	L.CheckType(-1, lua.LUA_TNUMBER)
	L.Pop(1)

	// Test CheckType with string
	L.PushString("test")
	L.CheckType(-1, lua.LUA_TSTRING)
	L.Pop(1)

	// Test CheckType with boolean
	L.PushBoolean(true)
	L.CheckType(-1, lua.LUA_TBOOLEAN)
	L.Pop(1)

	// Test CheckType with nil
	L.PushNil()
	L.CheckType(-1, lua.LUA_TNIL)
	L.Pop(1)

	// Test CheckType with function
	L.PushCFunction(func(L *lua.State) int { return 0 })
	L.CheckType(-1, lua.LUA_TFUNCTION)
	L.Pop(1)

	// Note: CheckType should panic/error with wrong types
	// but we can't easily test that in this test framework
}

func (s *Suite) TestCheckAny(assert *require.Assertions, L *lua.State) {
	// Test CheckAny with number
	L.PushNumber(42.5)
	// This should not panic/error as there is a value
	L.CheckAny(-1)
	L.Pop(1)

	// Test CheckAny with string
	L.PushString("test")
	L.CheckAny(-1)
	L.Pop(1)

	// Test CheckAny with boolean
	L.PushBoolean(false)
	L.CheckAny(-1)
	L.Pop(1)

	// Test CheckAny with nil
	L.PushNil()
	L.CheckAny(-1)
	L.Pop(1)

	// Test CheckAny with function
	L.PushCFunction(func(L *lua.State) int { return 0 })
	L.CheckAny(-1)
	L.Pop(1)

	L.AtPanic(func(L *lua.State) int {
		err := L.PopError()
		assert.Error(err)

		panic(err)
	})
	// Test CheckAny with invalid index (should panic)
	assert.Panics(func() {
		L.CheckAny(1)
	})
}

func (s *Suite) TestOptNumber(assert *require.Assertions, L *lua.State) {
	// Test OptNumber with valid number
	L.PushNumber(42.5)
	result := L.OptNumber(-1, 100.0)
	assert.Equal(42.5, result)
	L.Pop(1)

	// Test OptNumber with integer (should work as integers are numbers)
	L.PushInteger(123)
	result = L.OptNumber(-1, 100.0)
	assert.Equal(123.0, result)
	L.Pop(1)

	// Test OptNumber with string number (should work as Lua coerces)
	L.PushString("456.7")
	result = L.OptNumber(-1, 100.0)
	assert.Equal(456.7, result)
	L.Pop(1)

	// Test OptNumber with nil - should return default value
	L.PushNil()
	result = L.OptNumber(-1, 999.99)
	assert.Equal(999.99, result)
	L.Pop(1)

	// Test OptNumber with zero
	L.PushNumber(0.0)
	result = L.OptNumber(-1, 100.0)
	assert.Equal(0.0, result)
	L.Pop(1)

	// Test OptNumber with negative number
	L.PushNumber(-123.45)
	result = L.OptNumber(-1, 100.0)
	assert.Equal(-123.45, result)
	L.Pop(1)

	// Test OptNumber with invalid index (beyond stack) - should return default
	result = L.OptNumber(100, 777.77) // invalid index
	assert.Equal(777.77, result)

	// Test OptNumber with negative default
	L.PushNil()
	result = L.OptNumber(-1, -500.5)
	assert.Equal(-500.5, result)
	L.Pop(1)
}

func (s *Suite) TestOptInteger(assert *require.Assertions, L *lua.State) {
	// Test OptInteger with valid integer
	L.PushInteger(42)
	result := L.OptInteger(-1, 100)
	assert.Equal(int64(42), result)
	L.Pop(1)

	// Test OptInteger with exact floating point number (should work)
	L.PushNumber(123.0)
	result = L.OptInteger(-1, 100)
	assert.Equal(int64(123), result)
	L.Pop(1)

	// Test OptInteger with string integer (should work as Lua coerces)
	L.PushString("456")
	result = L.OptInteger(-1, 100)
	assert.Equal(int64(456), result)
	L.Pop(1)

	// Test OptInteger with nil - should return default value
	L.PushNil()
	result = L.OptInteger(-1, 999)
	assert.Equal(int64(999), result)
	L.Pop(1)

	// Test OptInteger with zero
	L.PushInteger(0)
	result = L.OptInteger(-1, 100)
	assert.Equal(int64(0), result)
	L.Pop(1)

	// Test OptInteger with negative integer
	L.PushInteger(-789)
	result = L.OptInteger(-1, 100)
	assert.Equal(int64(-789), result)
	L.Pop(1)

	// Test OptInteger with large integer
	L.PushInteger(math.MaxInt64) // max int64
	result = L.OptInteger(-1, 100)
	assert.Equal(int64(math.MaxInt64), result)
	L.Pop(1)

	// Test OptInteger with invalid index (beyond stack) - should return default
	result = L.OptInteger(100, 777) // invalid index
	assert.Equal(int64(777), result)

	// Test OptInteger with negative default
	L.PushNil()
	result = L.OptInteger(-1, -500)
	assert.Equal(int64(-500), result)
	L.Pop(1)
}

func (s *Suite) TestOptLString(assert *require.Assertions, L *lua.State) {
	// Test OptLString with valid string
	testStr := "Hello, World!"
	L.PushString(testStr)
	var size int
	result, err := L.OptLString(-1, "default", &size)
	assert.NoError(err)
	assert.Equal(testStr, result)
	assert.Equal(len(testStr), size)
	L.Pop(1)

	// Test OptLString with empty string
	L.PushString("")
	result, err = L.OptLString(-1, "default", &size)
	assert.NoError(err)
	assert.Equal("", result)
	assert.Equal(0, size)
	L.Pop(1)

	// Test OptLString with Unicode string
	unicodeStr := "你好世界🌍"
	L.PushString(unicodeStr)
	result, err = L.OptLString(-1, "default", &size)
	assert.NoError(err)
	assert.Equal(unicodeStr, result)
	assert.Equal(len(unicodeStr), size) // byte length
	L.Pop(1)

	// Test OptLString with number (should be coerced to string)
	L.PushNumber(42.5)
	result, err = L.OptLString(-1, "default", &size)
	assert.NoError(err)
	assert.Equal("42.5", result)
	assert.True(size > 0)
	L.Pop(1)

	// Test OptLString with integer (should be coerced to string)
	L.PushInteger(123)
	result, err = L.OptLString(-1, "default", &size)
	assert.NoError(err)
	assert.Equal("123", result)
	assert.Equal(3, size)
	L.Pop(1)

	// Test OptLString with nil - should return default value
	L.PushNil()
	defaultStr := "this is default"
	result, err = L.OptLString(-1, defaultStr, &size)
	assert.NoError(err)
	assert.Equal(defaultStr, result)
	assert.Equal(len(defaultStr), size)
	L.Pop(1)

	// Test OptLString with invalid index (beyond stack) - should return default
	defaultStr2 := "another default"
	result, err = L.OptLString(100, defaultStr2, &size) // invalid index
	assert.NoError(err)
	assert.Equal(defaultStr2, result)
	assert.Equal(len(defaultStr2), size)

	// Test OptLString with nil size pointer
	L.PushString("test")
	result, err = L.OptLString(-1, "default", nil)
	assert.NoError(err)
	assert.Equal("test", result)
	L.Pop(1)

	// Test OptLString with empty default
	L.PushNil()
	result, err = L.OptLString(-1, "", &size)
	assert.NoError(err)
	assert.Equal("", result)
	assert.Equal(0, size)
	L.Pop(1)

	// Test OptLString with long strings
	longStr := "This is a very long string that contains many characters and should test the string handling properly in the Lua to Go conversion"
	L.PushString(longStr)
	result, err = L.OptLString(-1, "default", &size)
	assert.NoError(err)
	assert.Equal(longStr, result)
	assert.Equal(len(longStr), size)
	L.Pop(1)
}
