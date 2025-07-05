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
