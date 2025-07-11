package lua_test

import (
	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/lua/lua54"
)

func (s *Suite) TestTable(assert *require.Assertions, L *lua.State) {
	L.CreateTable(2, 1)
	assert.Equal(lua.LUA_TTABLE, L.Type(-1))
	L.Pop(1)

	L.NewTable()
	assert.Equal(lua.LUA_TTABLE, L.Type(-1))

	L.PushString("hello")
	L.SetField(-2, "greeting")

	L.PushInteger(42)
	L.SetField(-2, "answer")

	typ, err := L.GetField(-1, "greeting")
	assert.NoError(err)
	assert.Equal(lua.LUA_TSTRING, typ)
	assert.True(L.IsString(-1))
	greeting := L.ToString(-1)
	assert.Equal("hello", greeting)
	L.Pop(1)

	typ, err = L.GetField(-1, "answer")
	assert.NoError(err)
	assert.Equal(lua.LUA_TNUMBER, typ)
	assert.True(L.IsInteger(-1))
	answer := L.ToInteger(-1)
	assert.Equal(int64(42), answer)
	L.Pop(1)

	typ, err = L.GetField(-1, "non_existent")
	assert.NoError(err)
	assert.Equal(lua.LUA_TNIL, typ)
	assert.True(L.IsNil(-1))
	L.Pop(1)

	L.PushString("first")
	L.SetI(-2, 1)

	L.PushString("second")
	L.SetI(-2, 2)

	assert.Equal(lua.LUA_TSTRING, L.GetI(-1, 1))
	assert.True(L.IsString(-1))
	first := L.ToString(-1)
	assert.Equal("first", first)
	L.Pop(1)

	assert.Equal(lua.LUA_TSTRING, L.GetI(-1, 2))
	assert.True(L.IsString(-1))
	second := L.ToString(-1)
	assert.Equal("second", second)
	L.Pop(1)

	assert.Equal(lua.LUA_TNIL, L.GetI(-1, 10))
	assert.True(L.IsNil(-1))
	L.Pop(1)

	L.PushString("stack_key")
	L.PushString("stack_value")
	L.SetTable(-3)

	L.PushString("stack_key")
	L.GetTable(-2)
	assert.True(L.IsString(-1))
	stackValue := L.ToString(-1)
	assert.Equal("stack_value", stackValue)
	L.Pop(1)

	L.PushString("bool_key")
	L.PushBoolean(true)
	L.SetTable(-3)

	L.PushString("bool_key")
	L.GetTable(-2)
	assert.True(L.IsBoolean(-1))
	boolVal := L.ToBoolean(-1)
	assert.True(boolVal)
	L.Pop(1)

	L.NewTable()
	L.PushString("inner_value")
	L.SetField(-2, "inner_key")
	L.SetField(-2, "nested")

	typ, err = L.GetField(-1, "nested")
	assert.NoError(err)
	assert.Equal(lua.LUA_TTABLE, typ)
	assert.Equal(lua.LUA_TTABLE, L.Type(-1))
	typ, err = L.GetField(-1, "inner_key")
	assert.NoError(err)
	assert.Equal(lua.LUA_TSTRING, typ)
	assert.True(L.IsString(-1))
	innerValue := L.ToString(-1)
	assert.Equal("inner_value", innerValue)
	L.Pop(2)

	L.PushString("new_greeting")
	L.SetField(-2, "greeting")

	typ, err = L.GetField(-1, "greeting")
	assert.NoError(err)
	assert.Equal(lua.LUA_TSTRING, typ)
	newGreeting := L.ToString(-1)
	assert.Equal("new_greeting", newGreeting)
	L.Pop(1)

	L.PushNil()
	L.SetField(-2, "answer")

	typ, err = L.GetField(-1, "answer")
	assert.NoError(err)
	assert.Equal(lua.LUA_TNIL, typ)
	assert.True(L.IsNil(-1))
	L.Pop(1)

	L.NewTable()
	for i := 1; i <= 5; i++ {
		L.PushInteger(int64(i * 10))
		L.SetI(-2, int64(i))
	}

	for i := 1; i <= 5; i++ {
		assert.Equal(lua.LUA_TNUMBER, L.GetI(-1, int64(i)))
		assert.True(L.IsInteger(-1))
		val := L.ToInteger(-1)
		assert.Equal(int64(i*10), val)
		L.Pop(1)
	}

	L.Pop(1)
	L.Pop(1)

	assert.Equal(0, L.GetTop())
}

func (s *Suite) TestTableRaw(assert *require.Assertions, L *lua.State) {
}
