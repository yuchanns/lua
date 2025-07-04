package lua_test

import (
	"fmt"
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
