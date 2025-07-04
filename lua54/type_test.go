package lua_test

import (
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
