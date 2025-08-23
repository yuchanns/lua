package lua_test

import (
	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/lua"
)

func (s *Suite) TestStackBasicOperations(assert *require.Assertions, L *lua.State) {

	assert.Equal(0, L.GetTop())

	L.PushInteger(42)
	L.PushString("hello")
	L.PushBoolean(true)

	assert.Equal(3, L.GetTop())

	L.SetTop(2)

	assert.Equal(2, L.GetTop())

	L.SetTop(5)

	assert.Equal(5, L.GetTop())

	assert.Equal(int64(42), L.ToInteger(1))
	assert.Equal("hello", L.ToString(2))
	assert.True(L.IsNil(3))
	assert.True(L.IsNil(4))
	assert.True(L.IsNil(5))

	L.SetTop(0)
	assert.Equal(0, L.GetTop())
}

func (s *Suite) TestStackPushValue(assert *require.Assertions, L *lua.State) {

	L.PushInteger(100)
	L.PushString("world")
	L.PushBoolean(true)

	assert.Equal(3, L.GetTop())

	L.PushValue(1)

	assert.Equal(4, L.GetTop())
	assert.Equal(int64(100), L.ToInteger(4))
	assert.Equal(int64(100), L.ToInteger(1))

	L.PushValue(-2)

	assert.Equal(5, L.GetTop())
	assert.Equal(true, L.ToBoolean(5))
	assert.Equal(true, L.ToBoolean(3))

	L.PushValue(2)

	assert.Equal(6, L.GetTop())
	assert.Equal("world", L.ToString(6))
	assert.Equal("world", L.ToString(2))
}

func (s *Suite) TestStackAbsIndex(assert *require.Assertions, L *lua.State) {

	L.PushInteger(1)
	L.PushInteger(2)
	L.PushInteger(3)
	L.PushInteger(4)

	stackSize := L.GetTop()
	assert.Equal(4, stackSize)

	assert.Equal(1, L.AbsIndex(1))
	assert.Equal(2, L.AbsIndex(2))
	assert.Equal(3, L.AbsIndex(3))
	assert.Equal(4, L.AbsIndex(4))

	assert.Equal(4, L.AbsIndex(-1))
	assert.Equal(3, L.AbsIndex(-2))
	assert.Equal(2, L.AbsIndex(-3))
	assert.Equal(1, L.AbsIndex(-4))

	L.SetTop(2)

	assert.Equal(2, L.AbsIndex(-1))
	assert.Equal(1, L.AbsIndex(-2))
}

func (s *Suite) TestStackCheckStack(assert *require.Assertions, L *lua.State) {

	assert.True(L.CheckStack(10))
	assert.True(L.CheckStack(100))
	assert.True(L.CheckStack(1000))

	for i := range 100 {
		L.PushInteger(int64(i))
	}

	assert.True(L.CheckStack(10))
	assert.True(L.CheckStack(100))
}

func (s *Suite) TestStackPop(assert *require.Assertions, L *lua.State) {

	L.PushInteger(1)
	L.PushInteger(2)
	L.PushInteger(3)
	L.PushInteger(4)
	L.PushInteger(5)

	assert.Equal(5, L.GetTop())

	L.Pop(1)
	assert.Equal(4, L.GetTop())
	assert.Equal(int64(4), L.ToInteger(-1))

	L.Pop(2)
	assert.Equal(2, L.GetTop())
	assert.Equal(int64(2), L.ToInteger(-1))

	L.Pop(0)
	assert.Equal(2, L.GetTop())

	L.Pop(2)
	assert.Equal(0, L.GetTop())
}

func (s *Suite) TestStackCopy(assert *require.Assertions, L *lua.State) {

	L.PushInteger(10)
	L.PushString("original")
	L.PushBoolean(false)
	L.PushInteger(20)

	assert.Equal(4, L.GetTop())

	L.Copy(1, 3)

	assert.Equal(4, L.GetTop())
	assert.Equal(int64(10), L.ToInteger(1))
	assert.Equal("original", L.ToString(2))
	assert.Equal(int64(10), L.ToInteger(3))
	assert.Equal(int64(20), L.ToInteger(4))

	L.Copy(-1, 2)

	assert.Equal(int64(20), L.ToInteger(2))
	assert.Equal(int64(20), L.ToInteger(-1))
}

func (s *Suite) TestStackRotate(assert *require.Assertions, L *lua.State) {

	for i := 1; i <= 5; i++ {
		L.PushInteger(int64(i))
	}

	L.Rotate(2, 1)

	assert.Equal(int64(1), L.ToInteger(1))
	assert.Equal(int64(5), L.ToInteger(2))
	assert.Equal(int64(2), L.ToInteger(3))
	assert.Equal(int64(3), L.ToInteger(4))
	assert.Equal(int64(4), L.ToInteger(5))

	L.SetTop(0)
	for i := 1; i <= 5; i++ {
		L.PushInteger(int64(i))
	}

	L.Rotate(2, -1)

	assert.Equal(int64(1), L.ToInteger(1))
	assert.Equal(int64(3), L.ToInteger(2))
	assert.Equal(int64(4), L.ToInteger(3))
	assert.Equal(int64(5), L.ToInteger(4))
	assert.Equal(int64(2), L.ToInteger(5))
}

func (s *Suite) TestStackInsert(assert *require.Assertions, L *lua.State) {

	L.PushInteger(10)
	L.PushInteger(20)
	L.PushInteger(30)

	L.PushInteger(99)
	assert.Equal(4, L.GetTop())

	L.Insert(2)
	assert.Equal(4, L.GetTop())

	assert.Equal(int64(10), L.ToInteger(1))
	assert.Equal(int64(20), L.ToInteger(3))
	assert.Equal(int64(30), L.ToInteger(4))
	assert.Equal(int64(99), L.ToInteger(2))

	L.PushInteger(77)
	L.Insert(1)
	assert.Equal(5, L.GetTop())

	assert.Equal(int64(77), L.ToInteger(1))
	assert.Equal(int64(10), L.ToInteger(2))
	assert.Equal(int64(99), L.ToInteger(3))
	assert.Equal(int64(20), L.ToInteger(4))
	assert.Equal(int64(30), L.ToInteger(5))
}

func (s *Suite) TestStackRemove(assert *require.Assertions, L *lua.State) {

	for i := 1; i <= 5; i++ {
		L.PushInteger(int64(i))
	}

	assert.Equal(5, L.GetTop())

	L.Remove(3)
	assert.Equal(4, L.GetTop())

	assert.Equal(int64(1), L.ToInteger(1))
	assert.Equal(int64(2), L.ToInteger(2))
	assert.Equal(int64(4), L.ToInteger(3))
	assert.Equal(int64(5), L.ToInteger(4))

	L.Remove(1)
	assert.Equal(3, L.GetTop())

	assert.Equal(int64(2), L.ToInteger(1))
	assert.Equal(int64(4), L.ToInteger(2))
	assert.Equal(int64(5), L.ToInteger(3))
}

func (s *Suite) TestStackReplace(assert *require.Assertions, L *lua.State) {

	for i := 1; i <= 4; i++ {
		L.PushInteger(int64(i))
	}

	L.PushInteger(99)
	assert.Equal(5, L.GetTop())

	L.Replace(2)
	assert.Equal(4, L.GetTop())

	assert.Equal(int64(1), L.ToInteger(1))
	assert.Equal(int64(99), L.ToInteger(2))
	assert.Equal(int64(3), L.ToInteger(3))
	assert.Equal(int64(4), L.ToInteger(4))

	L.PushInteger(77)
	L.Replace(1)
	assert.Equal(4, L.GetTop())

	assert.Equal(int64(77), L.ToInteger(1))
	assert.Equal(int64(99), L.ToInteger(2))
	assert.Equal(int64(3), L.ToInteger(3))
	assert.Equal(int64(4), L.ToInteger(4))
}

func (s *Suite) TestStackXMove(assert *require.Assertions, L *lua.State) {
	co := L.NewThread()
	L.SetTop(0)

	L.PushInteger(1)
	L.PushInteger(2)
	L.PushInteger(3)
	assert.Equal(3, L.GetTop())
	assert.Equal(0, co.GetTop())

	L.XMove(co, 2)
	assert.Equal(1, L.GetTop())
	assert.Equal(2, co.GetTop())

	assert.Equal(int64(1), L.ToInteger(1))

	assert.Equal(int64(2), co.ToInteger(1))
	assert.Equal(int64(3), co.ToInteger(2))

	L.XMove(co, 0)
	assert.Equal(1, L.GetTop())
	assert.Equal(2, co.GetTop())

	L.XMove(co, 1)
	assert.Equal(0, L.GetTop())
	assert.Equal(3, co.GetTop())

	assert.Equal(int64(2), co.ToInteger(1))
	assert.Equal(int64(3), co.ToInteger(2))
	assert.Equal(int64(1), co.ToInteger(3))
}

func (s *Suite) TestStackComplexOperations(assert *require.Assertions, L *lua.State) {

	L.PushInteger(1)
	L.PushInteger(2)
	L.PushInteger(3)
	L.PushInteger(4)
	L.PushInteger(5)

	L.PushValue(-1)

	assert.Equal(6, L.GetTop())
	assert.Equal(int64(5), L.ToInteger(-1))
	assert.Equal(int64(5), L.ToInteger(-2))

	L.Insert(3)

	assert.Equal(6, L.GetTop())

	assert.Equal(int64(1), L.ToInteger(1))
	assert.Equal(int64(2), L.ToInteger(2))
	assert.Equal(int64(5), L.ToInteger(3))
	assert.Equal(int64(3), L.ToInteger(4))
	assert.Equal(int64(4), L.ToInteger(5))
	assert.Equal(int64(5), L.ToInteger(6))

	L.Remove(3)
	assert.Equal(5, L.GetTop())

	for i := 1; i <= 5; i++ {
		assert.Equal(int64(i), L.ToInteger(i))
	}

	L.PushInteger(99)
	L.Replace(3)
	assert.Equal(5, L.GetTop())

	assert.Equal(int64(1), L.ToInteger(1))
	assert.Equal(int64(2), L.ToInteger(2))
	assert.Equal(int64(99), L.ToInteger(3))
	assert.Equal(int64(4), L.ToInteger(4))
	assert.Equal(int64(5), L.ToInteger(5))
}

func (s *Suite) TestStackRawEqual(assert *require.Assertions, L *lua.State) {
	L.PushInteger(123)
	L.PushInteger(123)
	assert.True(L.RawEqual(-1, -2))

	L.PushInteger(456)
	assert.False(L.RawEqual(-1, -2))

	L.SetTop(0)

	L.PushString("abc")
	L.PushString("abc")
	assert.True(L.RawEqual(-1, -2))

	L.PushString("def")
	assert.False(L.RawEqual(-1, -2))

	L.SetTop(0)

	L.PushBoolean(true)
	L.PushBoolean(true)
	assert.True(L.RawEqual(-1, -2))

	L.PushBoolean(false)
	assert.False(L.RawEqual(-1, -2))

	L.SetTop(0)

	L.PushNil()
	L.PushNil()
	assert.True(L.RawEqual(-1, -2))

	L.PushInteger(0)
	assert.False(L.RawEqual(-1, -2))
}

func (s *Suite) TestStackCompare(assert *require.Assertions, L *lua.State) {
	L.PushInteger(10)
	L.PushInteger(10)
	assert.True(L.Compare(-1, -2, lua.LUA_OPEQ))
	assert.False(L.Compare(-1, -2, lua.LUA_OPLT))
	assert.True(L.Compare(-1, -2, lua.LUA_OPLE))

	L.PushInteger(20)
	assert.False(L.Compare(-1, -2, lua.LUA_OPLT))
	assert.True(L.Compare(-2, -1, lua.LUA_OPLT))

	L.SetTop(0)

	L.PushString("foo")
	L.PushString("foo")
	assert.True(L.Compare(-1, -2, lua.LUA_OPEQ))

	L.PushString("zzz")
	assert.False(L.Compare(-1, -2, lua.LUA_OPLT))
	assert.True(L.Compare(-2, -1, lua.LUA_OPLT))

	L.SetTop(0)

	L.PushBoolean(true)
	L.PushBoolean(true)
	assert.True(L.Compare(-1, -2, lua.LUA_OPEQ))

	L.PushBoolean(false)
	assert.False(L.Compare(-1, -2, lua.LUA_OPEQ))

	L.SetTop(0)

	L.PushNil()
	L.PushNil()
	assert.True(L.Compare(-1, -2, lua.LUA_OPEQ))

	L.PushInteger(1)
	assert.False(L.Compare(-1, -2, lua.LUA_OPEQ))

	L.SetTop(0)
}

func (s *Suite) TestStackArith(assert *require.Assertions, L *lua.State) {
	L.PushInteger(15)
	L.PushInteger(27)
	L.Arith(lua.LUA_OPADD)
	assert.Equal(int64(42), L.ToInteger(-1))

	L.SetTop(0)

	L.PushNumber(100.5)
	L.PushNumber(0.5)
	L.Arith(lua.LUA_OPSUB)
	assert.InDelta(100.0, L.ToNumber(-1), 1e-7)

	L.SetTop(0)

	L.PushInteger(7)
	L.PushInteger(6)
	L.Arith(lua.LUA_OPMUL)
	assert.Equal(int64(42), L.ToInteger(-1))

	L.SetTop(0)

	L.PushInteger(126)
	L.PushInteger(3)
	L.Arith(lua.LUA_OPDIV)
	assert.InDelta(42.0, L.ToNumber(-1), 1e-7)

	L.SetTop(0)

	L.PushInteger(10)
	L.Arith(lua.LUA_OPUNM)
	assert.Equal(int64(-10), L.ToInteger(-1))

	L.SetTop(0)

	L.PushInteger(2)
	L.PushInteger(6)
	L.Arith(lua.LUA_OPPOW)
	assert.Equal(int64(64), L.ToInteger(-1))

	L.SetTop(0)
}

func (s *Suite) TestStackConcat(assert *require.Assertions, L *lua.State) {
	L.PushString("foo")
	L.PushString("bar")
	L.PushString("baz")
	L.Concat(3)
	assert.Equal("foobarbaz", L.ToString(-1))

	L.SetTop(0)

	L.PushInteger(123)
	L.PushString("hi-")
	L.Concat(2)
	assert.Equal("123hi-", L.ToString(-1))

	L.SetTop(0)

	L.PushString("hello")
	L.Concat(1)
	assert.Equal("hello", L.ToString(-1))

	L.SetTop(0)

	L.Concat(0)
	assert.Equal("", L.ToString(-1))
}

func (s *Suite) TestStackLen(assert *require.Assertions, L *lua.State) {
	L.PushString("foobar")
	L.Len(-1)
	assert.Equal(int64(6), L.ToInteger(-1))

	L.Pop(1)

	L.PushString("")
	L.Len(-1)
	assert.Equal(int64(0), L.ToInteger(-1))

	L.Pop(1)

	assert.NoError(L.LoadString("return {10, 20, 30, 40}"))
	L.Call(0, 1)
	L.Len(-1)
	assert.Equal(int64(4), L.ToInteger(-1))

	L.Pop(2)

	assert.NoError(L.LoadString("return {}"))
	L.Call(0, 1)
	L.Len(-1)
	assert.Equal(int64(0), L.ToInteger(-1))

	L.Pop(2)
}
