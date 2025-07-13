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

	assert.Error(err)
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
		return
	}
	return arena.ReAlloc(ptr, nsize)
}

func (s *Suite) TestLoad(assert *require.Assertions, L *lua.State) {

	script := "return 42"
	reader := strings.NewReader(script)
	err := L.Load(reader, "test_chunk")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)

	assert.True(L.IsNumber(-1))
	assert.Equal(42.0, L.ToNumber(-1))
	L.Pop(1)

	badScript := "return +"
	badReader := strings.NewReader(badScript)
	err = L.Load(badReader, "bad_chunk")
	assert.Error(err)

	script = "return 'text mode'"
	textReader := strings.NewReader(script)
	err = L.Load(textReader, "text_chunk", "t")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("text mode", L.ToString(-1))
	L.Pop(1)

	emptyReader := strings.NewReader("")
	err = L.Load(emptyReader, "empty_chunk")
	assert.NoError(err)
}

func (s *Suite) TestLoadBuffer(assert *require.Assertions, L *lua.State) {

	code := []byte("return 'hello from buffer'")
	err := L.LoadBuffer(code, "buffer_test")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("hello from buffer", L.ToString(-1))
	L.Pop(1)

	emptyCode := []byte("")
	err = L.LoadBuffer(emptyCode, "empty_buffer")
	assert.NoError(err)

	badCode := []byte("return {")
	err = L.LoadBuffer(badCode, "bad_buffer")
	assert.Error(err)

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

func (s *Suite) TestLoadBufferx(assert *require.Assertions, L *lua.State) {

	textCode := []byte("return 'text mode works'")
	err := L.LoadBufferx(textCode, "bufferx_text", "t")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("text mode works", L.ToString(-1))
	L.Pop(1)

	binaryCode := []byte("return 'binary mode'")
	err = L.LoadBufferx(binaryCode, "bufferx_binary", "bt")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("binary mode", L.ToString(-1))
	L.Pop(1)

	defaultCode := []byte("return 'default mode'")
	err = L.LoadBufferx(defaultCode, "bufferx_default")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("default mode", L.ToString(-1))
	L.Pop(1)

	validCode := []byte("return 42")
	err = L.LoadBufferx(validCode, "bufferx_invalid_mode", "xyz")

	if err == nil {
		err = L.PCall(0, 1, 0)
		if err == nil {
			L.Pop(1)
		}
	}

	syntaxError := []byte("return ]")
	err = L.LoadBufferx(syntaxError, "bufferx_syntax_error")
	assert.Error(err)
}

func (s *Suite) TestDoFile(assert *require.Assertions, L *lua.State) {

	err := L.DoFile("testdata/simple.lua")
	assert.NoError(err)

	assert.True(L.IsString(-1))
	assert.Equal("Hello from file!", L.ToString(-1))
	L.Pop(1)

	err = L.DoFile("testdata/function.lua")
	assert.NoError(err)

	assert.True(L.IsNumber(-1))
	assert.Equal(42.0, L.ToNumber(-1))
	L.Pop(1)

	err = L.DoFile("testdata/syntax_error.lua")
	assert.Error(err)

	err = L.DoFile("testdata/nonexistent.lua")
	assert.Error(err)

	err = L.DoFile("testdata/empty.lua")
	assert.NoError(err)
}

func (s *Suite) TestLoadFilex(assert *require.Assertions, L *lua.State) {

	err := L.LoadFilex("testdata/simple.lua", "t")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("Hello from file!", L.ToString(-1))
	L.Pop(1)

	err = L.LoadFilex("testdata/function.lua", "bt")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal(42.0, L.ToNumber(-1))
	L.Pop(1)

	err = L.LoadFilex("testdata/simple.lua")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("Hello from file!", L.ToString(-1))
	L.Pop(1)

	err = L.LoadFilex("testdata/syntax_error.lua")
	assert.Error(err)

	err = L.LoadFilex("testdata/nonexistent.lua")
	assert.Error(err)

	err = L.LoadFilex("testdata/empty.lua")
	assert.NoError(err)

	err = L.PCall(0, 0, 0)
	assert.NoError(err)
}

func (s *Suite) TestLoadFile(assert *require.Assertions, L *lua.State) {

	err := L.LoadFile("testdata/simple.lua")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("Hello from file!", L.ToString(-1))
	L.Pop(1)

	err = L.LoadFile("testdata/function.lua")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal(42.0, L.ToNumber(-1))
	L.Pop(1)

	err = L.LoadFile("testdata/syntax_error.lua")
	assert.Error(err)

	err = L.LoadFile("testdata/nonexistent.lua")
	assert.Error(err)

	err = L.LoadFile("testdata/empty.lua")
	assert.NoError(err)

	err = L.PCall(0, 0, 0)
	assert.NoError(err)

	err = L.LoadFile("testdata/simple.lua")
	assert.NoError(err)
	err = L.LoadFile("testdata/function.lua")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal(42.0, L.ToNumber(-1))
	L.Pop(1)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.Equal("Hello from file!", L.ToString(-1))
	L.Pop(1)
}

func (s *Suite) TestPCall(assert *require.Assertions, L *lua.State) {

	err := L.LoadString("return 42")
	assert.NoError(err)

	err = L.PCall(0, 1, 0)
	assert.NoError(err)
	assert.True(L.IsNumber(-1))
	assert.Equal(42.0, L.ToNumber(-1))
	L.Pop(1)

	err = L.LoadString("return function(a, b, c) return a + b, a * b, c end")
	assert.NoError(err)
	err = L.PCall(0, 1, 0)
	assert.NoError(err)

	L.PushNumber(10)
	L.PushNumber(20)
	L.PushString("test")

	err = L.PCall(3, 3, 0)
	assert.NoError(err)

	assert.Equal(30.0, L.ToNumber(-3))
	assert.Equal(200.0, L.ToNumber(-2))
	assert.Equal("test", L.ToString(-1))
	L.Pop(3)

	err = L.LoadString("error('test error')")
	assert.NoError(err)

	err = L.PCall(0, 0, 0)
	assert.Error(err)

	err = L.LoadString("return 1, 2, 3, 4, 5")
	assert.NoError(err)

	err = L.PCall(0, lua.LUA_MULTRET, 0)
	assert.NoError(err)

	for i := 1; i <= 5; i++ {
		assert.Equal(float64(i), L.ToNumber(-6+i))
	}
	L.Pop(5)

	L.PushCFunction(func(L *lua.State) int {
		assert.Equal("hello", L.ToString(1))
		return 0
	})
	L.SetGlobal("asserteq")
	err = L.LoadString("asserteq('hello')")
	assert.NoError(err)

	err = L.PCall(0, 0, 0)
	assert.NoError(err)

	err = L.LoadString("return function(err) return 'Handled: ' .. tostring(err) end")
	assert.NoError(err)
	err = L.PCall(0, 1, 0)
	assert.NoError(err)

	errHandlerIdx := L.GetTop()

	err = L.LoadString("error('original error')")
	assert.NoError(err)

	err = L.PCall(0, 1, errHandlerIdx)
	assert.Error(err)

	L.Pop(1)
}

func (s *Suite) TestCall(assert *require.Assertions, L *lua.State) {

	err := L.LoadString("return 42")
	assert.NoError(err)

	L.Call(0, 1)
	assert.True(L.IsNumber(-1))
	assert.Equal(42.0, L.ToNumber(-1))
	L.Pop(1)

	err = L.LoadString("return function(a, b, c) return a + b, a * b, c end")
	assert.NoError(err)
	L.Call(0, 1)

	L.PushNumber(10)
	L.PushNumber(20)
	L.PushString("test")

	L.Call(3, 3)

	assert.Equal(30.0, L.ToNumber(-3))
	assert.Equal(200.0, L.ToNumber(-2))
	assert.Equal("test", L.ToString(-1))
	L.Pop(3)

	err = L.LoadString("return 1, 2, 3, 4, 5")
	assert.NoError(err)

	L.Call(0, lua.LUA_MULTRET)

	for i := 1; i <= 5; i++ {
		assert.Equal(float64(i), L.ToNumber(-6+i))
	}
	L.Pop(5)

	L.PushCFunction(func(L *lua.State) int {
		assert.Equal("hello", L.ToString(1))
		return 0
	})
	L.SetGlobal("asserteq")
	err = L.LoadString("asserteq('hello')")
	assert.NoError(err)

	L.Call(0, 0)

	err = L.LoadString(`
		return function(x)
			local function double(n) return n * 2 end
			local function add_ten(n) return n + 10 end
			return add_ten(double(x))
		end
	`)
	assert.NoError(err)
	L.Call(0, 1)

	L.PushNumber(5)
	L.Call(1, 1)
	assert.Equal(20.0, L.ToNumber(-1))
	L.Pop(1)

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
	L.Call(0, 1)

	L.Call(0, 2)
	assert.Equal(6.0, L.ToNumber(-2))
	assert.Equal(0.0, L.ToNumber(-1))
	L.Pop(2)

	old := L.AtPanic(func(L *lua.State) int {
		msg := L.ToString(-1)
		panic(msg)
	})

	assert.Panics(func() {
		err := L.LoadString("error('test runtime error')")
		assert.NoError(err)
		L.Call(0, 0)
	})

	assert.Panics(func() {
		err := L.LoadString("local f = nil; f()")
		assert.NoError(err)
		L.Call(0, 0)
	})

	assert.Panics(func() {
		err := L.LoadString("error('division by zero')")
		assert.NoError(err)
		L.Call(0, 0)
	})

	L.AtPanic(old)

	err = L.LoadString("return 'normal execution'")
	assert.NoError(err)
	L.Call(0, 1)
	assert.Equal("normal execution", L.ToString(-1))
	L.Pop(1)
}
