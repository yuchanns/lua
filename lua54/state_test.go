package lua_test

import (
	"strings"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/lua/internal/tools"
	"go.yuchanns.xyz/lua/lua54"
)

func (s *Suite) TestAllocTracking(assert *require.Assertions, t *testing.T) {
	arena := tools.NewArena()
	t.Cleanup(arena.FreeAll)

	L, err := s.lib.NewState(lua.WithAlloc(trackingAlloc, arena))
	assert.NoError(err)

	t.Cleanup(L.Close)

	err = L.DoString(`local t = {}; for i=1,1000 do t[i] = i end`)
	assert.NoError(err)

	assert.NotZero(arena.TotalAllocated())
	assert.NotZero(arena.PeakMemory())
	assert.NotZero(arena.AllocCount())

	t.Logf("Total Allocated Memory: %d bytes", arena.TotalAllocated())
	t.Logf("Peak Memory Usage: %d bytes", arena.PeakMemory())
	t.Logf("Allocation Count: %d", arena.AllocCount())
}

func (s *Suite) TestAllocLimited(assert *require.Assertions, t *testing.T) {
	arena := tools.NewArena()
	t.Cleanup(arena.FreeAll)

	limitedMem := &limitedMemory{Limit: 1024 * 1024, Arena: arena}

	L, err := s.lib.NewState(lua.WithAlloc(limitedAlloc, limitedMem))
	assert.NoError(err)

	t.Cleanup(L.Close)

	err = L.DoString(`local t = {}; for i=1,100000 do t[i] = string.rep('x', 100) end`)

	assert.Error(err) // Should fail due to allocation limit
}

func trackingAlloc(arena *tools.Arena, ptr unsafe.Pointer, osize, nsize int) (newPtr unsafe.Pointer) {
	if nsize == 0 {
		arena.Free(ptr)
		return
	}
	return arena.ReAlloc(ptr, nsize)
}

type limitedMemory struct {
	Limit int
	Arena *tools.Arena
}

func limitedAlloc(limitedMem *limitedMemory, ptr unsafe.Pointer, osize, nsize int) (newPtr unsafe.Pointer) {
	arena := limitedMem.Arena
	if nsize == 0 {
		arena.Free(ptr)
		return
	}
	newUsed := arena.AllocCount() - osize + nsize
	if newUsed > limitedMem.Limit {
		return // Allocation exceeds limit
	}
	return arena.ReAlloc(ptr, nsize)
}

// Test Load method
func (s *Suite) TestLoad(assert *require.Assertions, L *lua.State) {
	// Test valid Lua script
	script := "return 42"
	reader := strings.NewReader(script)
	err := L.Load(reader, "test_chunk")
	assert.NoError(err)

	// Execute the loaded chunk
	err = L.PCall(0, 1, 0)
	assert.NoError(err)

	// Check the result
	assert.True(L.IsNumber(-1))
	assert.Equal(42.0, L.ToNumber(-1))
	L.Pop(1)

	// Test Load with syntax error
	badScript := "return +" // Invalid syntax
	badReader := strings.NewReader(badScript)
	err = L.Load(badReader, "bad_chunk")
	assert.Error(err)

	// Test Load with different mode
	script = "return 'text mode'"
	textReader := strings.NewReader(script)
	err = L.Load(textReader, "text_chunk", "t") // text mode only
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("text mode", L.ToString(-1))
	L.Pop(1)

	// Test Load with empty reader
	emptyReader := strings.NewReader("")
	err = L.Load(emptyReader, "empty_chunk")
	assert.NoError(err)
}

// Test LoadBuffer method
func (s *Suite) TestLoadBuffer(assert *require.Assertions, L *lua.State) {
	// Test with valid Lua code
	code := []byte("return 'hello from buffer'")
	err := L.LoadBuffer(code, "buffer_test")
	assert.NoError(err)

	// Execute the loaded buffer
	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("hello from buffer", L.ToString(-1))
	L.Pop(1)

	// Test with empty buffer
	emptyCode := []byte("")
	err = L.LoadBuffer(emptyCode, "empty_buffer")
	assert.NoError(err)

	// Test with syntax error
	badCode := []byte("return {") // Invalid syntax
	err = L.LoadBuffer(badCode, "bad_buffer")
	assert.Error(err)

	// Test with complex code
	complexCode := []byte(`
		local function add(a, b)
			return a + b
		end
		return add(10, 20)
	`)
	err = L.LoadBuffer(complexCode, "complex_buffer")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal(30.0, L.ToNumber(-1))
	L.Pop(1)
}

// Test LoadBufferx method
func (s *Suite) TestLoadBufferx(assert *require.Assertions, L *lua.State) {
	// Test with text mode
	textCode := []byte("return 'text mode works'")
	err := L.LoadBufferx(textCode, "bufferx_text", "t")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("text mode works", L.ToString(-1))
	L.Pop(1)

	// Test with binary mode (should work with text too)
	binaryCode := []byte("return 'binary mode'")
	err = L.LoadBufferx(binaryCode, "bufferx_binary", "bt")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("binary mode", L.ToString(-1))
	L.Pop(1)

	// Test with default mode (empty string)
	defaultCode := []byte("return 'default mode'")
	err = L.LoadBufferx(defaultCode, "bufferx_default")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("default mode", L.ToString(-1))
	L.Pop(1)

	// Test with invalid mode - this might not cause immediate error but could affect loading
	validCode := []byte("return 42")
	err = L.LoadBufferx(validCode, "bufferx_invalid_mode", "xyz")
	// The behavior depends on Lua version, might succeed or fail
	if err == nil {
		err = L.PCall(0, 1, 0)
		if err == nil {
			L.Pop(1) // Clean up if successful
		}
	}

	// Test syntax error
	syntaxError := []byte("return ]") // Invalid syntax
	err = L.LoadBufferx(syntaxError, "bufferx_syntax_error")
	assert.Error(err)
}

// Test DoFile method
func (s *Suite) TestDoFile(assert *require.Assertions, L *lua.State) {
	// Test DoFile with simple file
	err := L.DoFile("testdata/simple.lua")
	assert.NoError(err)

	// Check the result
	assert.True(L.IsString(-1))
	assert.Equal("Hello from file!", L.ToString(-1))
	L.Pop(1)

	// Test DoFile with function file
	err = L.DoFile("testdata/function.lua")
	assert.NoError(err)

	assert.True(L.IsNumber(-1))
	assert.Equal(42.0, L.ToNumber(-1))
	L.Pop(1)

	// Test DoFile with syntax error file
	err = L.DoFile("testdata/syntax_error.lua")
	assert.Error(err)

	// Test DoFile with non-existent file
	err = L.DoFile("testdata/nonexistent.lua")
	assert.Error(err)

	// Test DoFile with empty file
	err = L.DoFile("testdata/empty.lua")
	assert.NoError(err) // Empty file should be OK, just no return value
}

// Test LoadFilex method
func (s *Suite) TestLoadFilex(assert *require.Assertions, L *lua.State) {
	// Test LoadFilex with text mode
	err := L.LoadFilex("testdata/simple.lua", "t")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("Hello from file!", L.ToString(-1))
	L.Pop(1)

	// Test LoadFilex with binary mode
	err = L.LoadFilex("testdata/function.lua", "bt")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal(42.0, L.ToNumber(-1))
	L.Pop(1)

	// Test LoadFilex with default mode (empty string)
	err = L.LoadFilex("testdata/simple.lua")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("Hello from file!", L.ToString(-1))
	L.Pop(1)

	// Test LoadFilex with syntax error
	err = L.LoadFilex("testdata/syntax_error.lua")
	assert.Error(err)

	// Test LoadFilex with non-existent file
	err = L.LoadFilex("testdata/nonexistent.lua")
	assert.Error(err)

	// Test LoadFilex with empty file
	err = L.LoadFilex("testdata/empty.lua")
	assert.NoError(err) // Should load but not execute anything

	err = L.PCall(0, 0, 0)
	assert.NoError(err)
}

// Test LoadFile method
func (s *Suite) TestLoadFile(assert *require.Assertions, L *lua.State) {
	// Test LoadFile with simple file
	err := L.LoadFile("testdata/simple.lua")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("Hello from file!", L.ToString(-1))
	L.Pop(1)

	// Test LoadFile with function file
	err = L.LoadFile("testdata/function.lua")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal(42.0, L.ToNumber(-1))
	L.Pop(1)

	// Test LoadFile with syntax error
	err = L.LoadFile("testdata/syntax_error.lua")
	assert.Error(err)

	// Test LoadFile with non-existent file
	err = L.LoadFile("testdata/nonexistent.lua")
	assert.Error(err)

	// Test LoadFile with empty file
	err = L.LoadFile("testdata/empty.lua")
	assert.NoError(err)

	err = L.PCall(0, 0, 0)
	assert.NoError(err)

	// Test multiple LoadFile calls
	err = L.LoadFile("testdata/simple.lua")
	assert.NoError(err)
	err = L.LoadFile("testdata/function.lua")
	assert.NoError(err)

	// Execute function.lua first (top of stack)
	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal(42.0, L.ToNumber(-1))
	L.Pop(1)

	// Execute simple.lua next
	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("Hello from file!", L.ToString(-1))
	L.Pop(1)
}

// Test PCall method
func (s *Suite) TestPCall(assert *require.Assertions, L *lua.State) {
	// Test PCall with successful function execution
	err := L.LoadString("return 42")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.True(L.IsNumber(-1))
	assert.Equal(42.0, L.ToNumber(-1))
	L.Pop(1)

	// Test PCall with multiple arguments and returns
	err = L.LoadString("return function(a, b, c) return a + b, a * b, c end")
	assert.NoError(err)
	err = L.PCall(0, 1, 0) // Load the function
	assert.NoError(err)

	// Push arguments
	L.PushNumber(10)
	L.PushNumber(20)
	L.PushString("test")

	// Call function with 3 args, expecting 3 returns
	err = L.PCall(3, 3, 0)
	assert.NoError(err)

	// Check results (stack order: bottom to top)
	assert.Equal(30.0, L.ToNumber(-3))   // a + b
	assert.Equal(200.0, L.ToNumber(-2))  // a * b
	assert.Equal("test", L.ToString(-1)) // c
	L.Pop(3)

	// Test PCall with runtime error
	err = L.LoadString("error('test error')")
	assert.NoError(err)

	err = L.PCall(0, 0, 0)
	assert.Error(err)

	// Test PCall with LUA_MULTRET
	err = L.LoadString("return 1, 2, 3, 4, 5")
	assert.NoError(err)

	err = L.PCall(0, lua.LUA_MULTRET, 0)
	assert.NoError(err)

	// Should have 5 values on stack
	for i := 1; i <= 5; i++ {
		assert.Equal(float64(i), L.ToNumber(-6+i))
	}
	L.Pop(5)

	// Test PCall with 0 returns
	L.PushGoFunction(func(msg string) {
		assert.Equal("hello", msg)
	})
	L.SetGlobal("asserteq")
	err = L.LoadString("asserteq('hello')")
	assert.NoError(err)

	err = L.PCall(0, 0, 0)
	assert.NoError(err)
	// No values should be left on stack

	// Test PCall with error handler function
	// First push error handler
	err = L.LoadString("return function(err) return 'Handled: ' .. tostring(err) end")
	assert.NoError(err)
	err = L.PCall(0, 1, 0)
	assert.NoError(err)

	errHandlerIdx := L.GetTop()

	// Push function that will error
	err = L.LoadString("error('original error')")
	assert.NoError(err)

	// Call with error handler
	err = L.PCall(0, 1, errHandlerIdx)
	assert.Error(err)

	L.Pop(1) // Pop error handler
}

// Test Call method
func (s *Suite) TestCall(assert *require.Assertions, L *lua.State) {
	// Test Call with successful function execution
	err := L.LoadString("return 42")
	assert.NoError(err)

	L.Call(0, 1)
	assert.True(L.IsNumber(-1))
	assert.Equal(42.0, L.ToNumber(-1))
	L.Pop(1)

	// Test Call with multiple arguments and returns
	err = L.LoadString("return function(a, b, c) return a + b, a * b, c end")
	assert.NoError(err)
	L.Call(0, 1) // Load the function
	
	// Push arguments
	L.PushNumber(10)
	L.PushNumber(20)
	L.PushString("test")

	// Call function with 3 args, expecting 3 returns
	L.Call(3, 3)

	// Check results (stack order: bottom to top)
	assert.Equal(30.0, L.ToNumber(-3))   // a + b
	assert.Equal(200.0, L.ToNumber(-2))  // a * b
	assert.Equal("test", L.ToString(-1)) // c
	L.Pop(3)

	// Test Call with LUA_MULTRET
	err = L.LoadString("return 1, 2, 3, 4, 5")
	assert.NoError(err)

	L.Call(0, lua.LUA_MULTRET)

	// Should have 5 values on stack
	for i := 1; i <= 5; i++ {
		assert.Equal(float64(i), L.ToNumber(-6+i))
	}
	L.Pop(5)

	// Test Call with 0 returns
	L.PushGoFunction(func(msg string) {
		assert.Equal("hello", msg)
	})
	L.SetGlobal("asserteq")
	err = L.LoadString("asserteq('hello')")
	assert.NoError(err)

	L.Call(0, 0)
	// No values should be left on stack

	// Test Call with nested function calls
	err = L.LoadString(`
		return function(x)
			local function double(n) return n * 2 end
			local function add_ten(n) return n + 10 end
			return add_ten(double(x))
		end
	`)
	assert.NoError(err)
	L.Call(0, 1) // Load the function

	L.PushNumber(5)
	L.Call(1, 1) // Call with 5, should return (5*2)+10 = 20
	assert.Equal(20.0, L.ToNumber(-1))
	L.Pop(1)

	// Test Call with table operations
	err = L.LoadString(`
		return function()
			local t = {a = 1, b = 2, c = 3}
			local sum = 0
			for k, v in pairs(t) do
				sum = sum + v
			end
			return sum, #t
		end
	`)
	assert.NoError(err)
	L.Call(0, 1) // Load the function

	L.Call(0, 2) // Should return sum=6 and length=0 (hash table has no array part)
	assert.Equal(6.0, L.ToNumber(-2))  // sum
	assert.Equal(0.0, L.ToNumber(-1))  // length of hash table
	L.Pop(2)

	// Note: We cannot test Call with runtime errors because Call will panic
	// instead of returning an error. That's the main difference from PCall.
	// Any runtime error in Call would cause the test to fail with panic.
}
