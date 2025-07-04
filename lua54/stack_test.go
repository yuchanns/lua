package lua_test

import (
	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/lua/lua54"
)

// Test basic stack operations: GetTop, SetTop
func (s *Suite) TestStackBasicOperations(assert *require.Assertions, L *lua.State) {
	// Test GetTop with empty stack
	assert.Equal(0, L.GetTop())

	// Push some values
	L.PushInteger(42)
	L.PushString("hello")
	L.PushBoolean(true)
	// Stack should now have 3 elements: [42, "hello", true]

	// Test GetTop
	assert.Equal(3, L.GetTop())

	// Test SetTop to reduce stack
	L.SetTop(2)
	// Stack should now have 2 elements: [42, "hello"]
	assert.Equal(2, L.GetTop())

	// Test SetTop to increase stack (fills with nil)
	L.SetTop(5)
	// Stack should now have 5 elements: [42, "hello", nil, nil, nil]
	assert.Equal(5, L.GetTop())

	// Verify the values we can access
	assert.Equal(int64(42), L.ToInteger(1))
	assert.Equal("hello", L.ToString(2))
	assert.True(L.IsNil(3))
	assert.True(L.IsNil(4))
	assert.True(L.IsNil(5))

	// Reset stack
	L.SetTop(0)
	assert.Equal(0, L.GetTop())
}

// Test PushValue operation
func (s *Suite) TestStackPushValue(assert *require.Assertions, L *lua.State) {
	// Push original values
	L.PushInteger(100)
	L.PushString("world")
	L.PushBoolean(true)

	assert.Equal(3, L.GetTop())

	// Test PushValue with positive index
	L.PushValue(1) // Copy integer at index 1 onto the top
	// Stack should now have 4 elements: [100, "world", true, 100]
	assert.Equal(4, L.GetTop())
	assert.Equal(int64(100), L.ToInteger(4))
	assert.Equal(int64(100), L.ToInteger(1)) // Original still there

	// Test PushValue with negative index
	L.PushValue(-2) // Copy boolean (index 3, which is -2 from top)
	// Stack should now have 5 elements: [100, "world", true, 100, true]
	assert.Equal(5, L.GetTop())
	assert.Equal(true, L.ToBoolean(5))
	assert.Equal(true, L.ToBoolean(3)) // Original still there

	// Test PushValue with string
	L.PushValue(2) // Copy string at index 2
	// Stack should now have 6 elements: [100, "world", true, 100, true, "world"]
	assert.Equal(6, L.GetTop())
	assert.Equal("world", L.ToString(6))
	assert.Equal("world", L.ToString(2)) // Original still there
}

// Test AbsIndex function
func (s *Suite) TestStackAbsIndex(assert *require.Assertions, L *lua.State) {
	// Push some values
	L.PushInteger(1)
	L.PushInteger(2)
	L.PushInteger(3)
	L.PushInteger(4)

	stackSize := L.GetTop()
	assert.Equal(4, stackSize)

	// Test positive indices (should remain unchanged)
	assert.Equal(1, L.AbsIndex(1))
	assert.Equal(2, L.AbsIndex(2))
	assert.Equal(3, L.AbsIndex(3))
	assert.Equal(4, L.AbsIndex(4))

	// Test negative indices (should convert to positive)
	assert.Equal(4, L.AbsIndex(-1)) // Last element
	assert.Equal(3, L.AbsIndex(-2)) // Second to last
	assert.Equal(2, L.AbsIndex(-3)) // Third to last
	assert.Equal(1, L.AbsIndex(-4)) // First element

	// Test with different stack sizes
	L.SetTop(2) // Reduce to 2 elements
	// Stack should now be: [1, 2]
	assert.Equal(2, L.AbsIndex(-1))
	assert.Equal(1, L.AbsIndex(-2))
}

// Test CheckStack function
func (s *Suite) TestStackCheckStack(assert *require.Assertions, L *lua.State) {
	// Should be able to check for reasonable stack space
	assert.True(L.CheckStack(10))
	assert.True(L.CheckStack(100))
	assert.True(L.CheckStack(1000))

	// Fill up some stack space
	for i := 0; i < 100; i++ {
		L.PushInteger(int64(i))
	}

	// Should still be able to allocate more
	assert.True(L.CheckStack(10))
	assert.True(L.CheckStack(100))
}

// Test Pop function
func (s *Suite) TestStackPop(assert *require.Assertions, L *lua.State) {
	// Push some values
	L.PushInteger(1)
	L.PushInteger(2)
	L.PushInteger(3)
	L.PushInteger(4)
	L.PushInteger(5)

	assert.Equal(5, L.GetTop())

	// Test Pop(1)
	L.Pop(1)
	assert.Equal(4, L.GetTop())
	assert.Equal(int64(4), L.ToInteger(-1)) // Top should now be 4

	// Test Pop(2)
	L.Pop(2)
	assert.Equal(2, L.GetTop())
	assert.Equal(int64(2), L.ToInteger(-1)) // Top should now be 2

	// Test Pop(0) - should do nothing
	L.Pop(0)
	assert.Equal(2, L.GetTop())

	// Pop remaining elements
	L.Pop(2)
	assert.Equal(0, L.GetTop())
}

// Test Copy function
func (s *Suite) TestStackCopy(assert *require.Assertions, L *lua.State) {
	// Push initial values
	L.PushInteger(10)
	L.PushString("original")
	L.PushBoolean(false)
	L.PushInteger(20)

	assert.Equal(4, L.GetTop())

	// Test Copy from index 1 to index 3
	L.Copy(1, 3) // Copy integer 10 to index 3 (overwriting boolean)
	// Stack should now be: [10, "original", 10, 20]
	assert.Equal(4, L.GetTop())             // Stack size unchanged
	assert.Equal(int64(10), L.ToInteger(1)) // Original still there
	assert.Equal("original", L.ToString(2)) // Unchanged
	assert.Equal(int64(10), L.ToInteger(3)) // Copied value
	assert.Equal(int64(20), L.ToInteger(4)) // Unchanged

	// Test Copy with negative indices
	L.Copy(-1, 2) // Copy top value (20) to index 2
	// Stack should now be: [10, 20, 10, 20]
	assert.Equal(int64(20), L.ToInteger(2))  // String replaced with integer
	assert.Equal(int64(20), L.ToInteger(-1)) // Original still at top
}

// Test Rotate function
func (s *Suite) TestStackRotate(assert *require.Assertions, L *lua.State) {
	// Setup: Push values 1, 2, 3, 4, 5
	for i := 1; i <= 5; i++ {
		L.PushInteger(int64(i))
	}

	// Test Rotate with positive n (rotate right)
	L.Rotate(2, 1) // Rotate elements from index 2 to top, 1 position right
	// Stack should now be: 1, 5, 2, 3, 4
	assert.Equal(int64(1), L.ToInteger(1))
	assert.Equal(int64(5), L.ToInteger(2))
	assert.Equal(int64(2), L.ToInteger(3))
	assert.Equal(int64(3), L.ToInteger(4))
	assert.Equal(int64(4), L.ToInteger(5))

	// Reset stack
	L.SetTop(0)
	for i := 1; i <= 5; i++ {
		L.PushInteger(int64(i))
	}

	// Test Rotate with negative n (rotate left)
	L.Rotate(2, -1) // Rotate elements from index 2 to top, 1 position left
	// Stack should now be: 1, 3, 4, 5, 2
	assert.Equal(int64(1), L.ToInteger(1))
	assert.Equal(int64(3), L.ToInteger(2))
	assert.Equal(int64(4), L.ToInteger(3))
	assert.Equal(int64(5), L.ToInteger(4))
	assert.Equal(int64(2), L.ToInteger(5))
}

// Test Insert function
func (s *Suite) TestStackInsert(assert *require.Assertions, L *lua.State) {
	// Push values: 10, 20, 30
	L.PushInteger(10)
	L.PushInteger(20)
	L.PushInteger(30)

	// Push value to insert: 99
	L.PushInteger(99)
	assert.Equal(4, L.GetTop())

	// Test Insert at index 2
	L.Insert(2)                 // Insert 99 at position 2
	assert.Equal(4, L.GetTop()) // Stack size unchanged

	// Stack should now be: 10, 99, 20, 30
	assert.Equal(int64(10), L.ToInteger(1))
	assert.Equal(int64(20), L.ToInteger(3))
	assert.Equal(int64(30), L.ToInteger(4))
	assert.Equal(int64(99), L.ToInteger(2))

	// Test Insert at index 1 (beginning)
	L.PushInteger(77)
	L.Insert(1)
	assert.Equal(5, L.GetTop())

	// Stack should now be: 77, 10, 99, 20, 30
	assert.Equal(int64(77), L.ToInteger(1))
	assert.Equal(int64(10), L.ToInteger(2))
	assert.Equal(int64(99), L.ToInteger(3))
	assert.Equal(int64(20), L.ToInteger(4))
	assert.Equal(int64(30), L.ToInteger(5))
}

// Test Remove function
func (s *Suite) TestStackRemove(assert *require.Assertions, L *lua.State) {
	// Push values: 1, 2, 3, 4, 5
	for i := 1; i <= 5; i++ {
		L.PushInteger(int64(i))
	}

	assert.Equal(5, L.GetTop())

	// Test Remove at index 3
	L.Remove(3)
	assert.Equal(4, L.GetTop())

	// Stack should now be: 1, 2, 4, 5
	assert.Equal(int64(1), L.ToInteger(1))
	assert.Equal(int64(2), L.ToInteger(2))
	assert.Equal(int64(4), L.ToInteger(3))
	assert.Equal(int64(5), L.ToInteger(4))

	// Test Remove at index 1 (first element)
	L.Remove(1)
	assert.Equal(3, L.GetTop())

	// Stack should now be: 2, 4, 5
	assert.Equal(int64(2), L.ToInteger(1))
	assert.Equal(int64(4), L.ToInteger(2))
	assert.Equal(int64(5), L.ToInteger(3))
}

// Test Replace function
func (s *Suite) TestStackReplace(assert *require.Assertions, L *lua.State) {
	// Push values: 1, 2, 3, 4
	for i := 1; i <= 4; i++ {
		L.PushInteger(int64(i))
	}

	// Push replacement value: 99
	L.PushInteger(99)
	assert.Equal(5, L.GetTop())

	// Test Replace at index 2
	L.Replace(2)                // Replace value at index 2 with 99
	assert.Equal(4, L.GetTop()) // Stack size reduced by 1

	// Stack should now be: 1, 99, 3, 4
	assert.Equal(int64(1), L.ToInteger(1))
	assert.Equal(int64(99), L.ToInteger(2))
	assert.Equal(int64(3), L.ToInteger(3))
	assert.Equal(int64(4), L.ToInteger(4))

	// Test Replace at index 1 (first element)
	L.PushInteger(77)
	L.Replace(1)
	assert.Equal(4, L.GetTop())

	// Stack should now be: 77, 99, 3, 4
	assert.Equal(int64(77), L.ToInteger(1))
	assert.Equal(int64(99), L.ToInteger(2))
	assert.Equal(int64(3), L.ToInteger(3))
	assert.Equal(int64(4), L.ToInteger(4))
}

// Test XMove function
func (s *Suite) TestStackXMove(assert *require.Assertions, L *lua.State) {
	// Create a second Lua state
	L2, err := s.lib.NewState()
	assert.NoError(err)
	defer L2.Close()

	// Push values in L: 1, 2, 3
	L.PushInteger(1)
	L.PushInteger(2)
	L.PushInteger(3)
	assert.Equal(3, L.GetTop())
	assert.Equal(0, L2.GetTop())

	// Test XMove: move 2 values from L to L2
	L.XMove(L2, 2)
	assert.Equal(1, L.GetTop())  // L should have 1 value left
	assert.Equal(2, L2.GetTop()) // L2 should have 2 values

	// Check values in L
	assert.Equal(int64(1), L.ToInteger(1))

	// Check values in L2 (order should be preserved)
	assert.Equal(int64(2), L2.ToInteger(1))
	assert.Equal(int64(3), L2.ToInteger(2))

	// Test XMove with 0 values
	L.XMove(L2, 0)
	assert.Equal(1, L.GetTop())
	assert.Equal(2, L2.GetTop())

	// Test XMove remaining value
	L.XMove(L2, 1)
	assert.Equal(0, L.GetTop())
	assert.Equal(3, L2.GetTop())

	// Check final state of L2
	assert.Equal(int64(2), L2.ToInteger(1))
	assert.Equal(int64(3), L2.ToInteger(2))
	assert.Equal(int64(1), L2.ToInteger(3))
}

// Test complex stack operations
func (s *Suite) TestStackComplexOperations(assert *require.Assertions, L *lua.State) {
	// Test a complex sequence of operations
	L.PushInteger(1)
	L.PushInteger(2)
	L.PushInteger(3)
	L.PushInteger(4)
	L.PushInteger(5)
	// Stack should now be: 1, 2, 3, 4, 5

	// Duplicate the top value
	L.PushValue(-1)
	// Stack should now be: 1, 2, 3, 4, 5, 5
	assert.Equal(6, L.GetTop())
	assert.Equal(int64(5), L.ToInteger(-1))
	assert.Equal(int64(5), L.ToInteger(-2))

	// Moves top value to position 3 and shifts up the elements above it
	L.Insert(3)
	// Stack should now be: 1, 2, 5, 3, 4, 5
	assert.Equal(6, L.GetTop())
	// Stack should now be: 1, 2, 5, 3, 4, 5
	assert.Equal(int64(1), L.ToInteger(1))
	assert.Equal(int64(2), L.ToInteger(2))
	assert.Equal(int64(5), L.ToInteger(3))
	assert.Equal(int64(3), L.ToInteger(4))
	assert.Equal(int64(4), L.ToInteger(5))
	assert.Equal(int64(5), L.ToInteger(6))

	// Remove the duplicate at position 3
	L.Remove(3)
	assert.Equal(5, L.GetTop())
	// Stack should now be: 1, 2, 3, 4, 5
	for i := 1; i <= 5; i++ {
		assert.Equal(int64(i), L.ToInteger(i))
	}

	// Replace position 3 with a new value
	L.PushInteger(99)
	L.Replace(3)
	assert.Equal(5, L.GetTop())
	// Stack should now be: 1, 2, 99, 4, 5
	assert.Equal(int64(1), L.ToInteger(1))
	assert.Equal(int64(2), L.ToInteger(2))
	assert.Equal(int64(99), L.ToInteger(3))
	assert.Equal(int64(4), L.ToInteger(4))
	assert.Equal(int64(5), L.ToInteger(5))
}

func (s *Suite) TestAtPanic(assert *require.Assertions, L *lua.State) {
	_ = L.AtPanic(func(L *lua.State) int {
		err := L.PopError()
		assert.Error(err)
		// We raise a panic to avoid SIGABRT abort the tests
		panic(err.Error())
	})

	errMsg := "Oops, no enclosing lua_pcall"
	assert.PanicsWithValue(
		errMsg,
		func() {
			L.Errorf("%s", errMsg)
		},
	)
}
