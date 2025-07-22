package lua_test

import (
	"unsafe"

	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/lua/lua54"
)

func (s *Suite) TestThread(assert *require.Assertions, L *lua.State) {
	co := L.NewThread()
	assert.NotNil(co)
	assert.Equal(0, co.GetTop())

	err := co.CloseThread(L)
	assert.NoError(err)

	assert.True(L.PushThread())
	L.Pop(1)

	assert.False(co.PushThread())
	assert.NoError(co.CloseThread(L))

	assert.Equal(lua.LUA_OK, co.Status())

	assert.False(L.IsYieldable())
	assert.True(co.IsYieldable())
}

func (s *Suite) TestThreadScript(assert *require.Assertions, L *lua.State) {
	L.DoFile("testdata/coro.lua")
	co := L.ToThread(-1)
	assert.NotNil(co)
	assert.Equal(lua.LUA_OK, co.Status())
	L.Pop(1)

	retc, err := co.Resume(L, 0)
	assert.NoError(err)
	assert.EqualValues(1, retc)
	assert.EqualValues(1, co.ToNumber(-1))
	co.Pop(1)

	retc, err = co.Resume(L, 0)
	assert.NoError(err)
	assert.EqualValues(1, retc)
	assert.EqualValues(2, co.ToNumber(-1))
	co.Pop(1)

	retc, err = co.Resume(L, 0)
	assert.NoError(err)
	assert.EqualValues(1, retc)
	assert.EqualValues(99, co.ToNumber(-1))
	co.Pop(1)

	_, err = co.Resume(L, 0)
	assert.Error(err)

	assert.NoError(co.CloseThread(L))

	err = co.DoFile("testdata/yield_and_sum.lua")
	assert.NoError(err)
	co.PushInteger(3)
	assert.Equal(lua.LUA_OK, co.Status())

	nres, err := co.Resume(L, 1)
	assert.NoError(err)
	assert.EqualValues(2, nres)
	assert.EqualValues(3, co.ToNumber(-2))
	assert.EqualValues(9, co.ToNumber(-1))
	co.Pop(2)

	nres, err = co.Resume(L, 0)
	assert.NoError(err)
	assert.EqualValues(1, nres)
	assert.EqualValues(6, co.ToNumber(-1))
	co.Pop(1)

	_, err = co.Resume(L, 0)
	assert.Error(err)
}

func (s *Suite) TestThreadYield(assert *require.Assertions, L *lua.State) {
	type fibContext struct {
		a int64
		b int64
		n int64
		i int64
	}
	var fibCont lua.KFunc
	var fc = &fibContext{}
	fibCont = func(L *lua.State, status int, ctx unsafe.Pointer) int {
		fc := (*fibContext)(ctx)
		if fc.i > fc.n {
			return 0
		}
		L.PushInteger(fc.a)
		fc.a, fc.b = fc.b, fc.a+fc.b
		fc.i++
		assert.NoError(L.YieldK(1, unsafe.Pointer(fc), fibCont))
		return 0
	}
	var fib lua.GoFunc = func(L *lua.State) int {
		n := L.CheckInteger(1)
		if n <= 0 {
			L.PushInteger(0)
			return 1
		}
		a, b := int64(0), int64(1)
		fc.a = b
		fc.b = a + b
		fc.n = n
		fc.i = 1
		L.PushInteger(a)
		assert.NoError(L.YieldK(1, unsafe.Pointer(fc), fibCont))
		return 0
	}
	L.PushGoFunction(fib)
	assert.NoError(L.SetGlobal("fib"))

	assert.NoError(L.DoFile("testdata/resume.lua"))
}
