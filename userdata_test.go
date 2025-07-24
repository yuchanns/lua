package lua_test

import (
	"unsafe"

	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/lua"
)

func (s *Suite) TestUserData(assert *require.Assertions, L *lua.State) {
	assert.NotNil(L.NewUserData(16))
	assert.Equal(lua.LUA_TUSERDATA, L.Type(-1))
	L.Pop(1)

	assert.NotNil(L.NewUserData(0))
	assert.Equal(lua.LUA_TUSERDATA, L.Type(-1))
	L.Pop(1)

	assert.NotNil(L.NewUserDataUv(8, 2))
	udIdx := L.GetTop()
	assert.Equal(lua.LUA_TUSERDATA, L.Type(-1))

	for i := 1; i <= 2; i++ {
		t := L.GetIUserValue(udIdx, i)
		assert.Equal(lua.LUA_TNIL, int(t))
		assert.True(L.IsNil(-1))
		L.Pop(1)
	}
	L.PushInteger(123)
	L.SetIUserValue(udIdx, 1)
	L.PushString("uvstr")
	L.SetIUserValue(udIdx, 2)
	L.PushInteger(10)
	L.SetIUserValue(udIdx, 1)

	t := L.GetIUserValue(udIdx, 1)
	assert.Equal(lua.LUA_TNUMBER, t)
	assert.True(L.IsInteger(-1))
	assert.Equal(int64(10), L.ToInteger(-1))
	L.Pop(1)

	t = L.GetIUserValue(udIdx, 2)
	assert.Equal(lua.LUA_TSTRING, t)
	assert.True(L.IsString(-1))
	assert.Equal("uvstr", L.ToString(-1))
	L.Pop(1)

	t = L.GetIUserValue(udIdx, 10)
	assert.Equal(lua.LUA_TNONE, t)
	assert.True(L.IsNil(-1))
	L.Pop(1)

	assert.NotNil(L.NewUserDataUv(16, 1))
	udIdxOne := L.GetTop()
	assert.Equal(lua.LUA_TUSERDATA, L.Type(-1))
	L.PushString("onlyuv")
	L.SetUserValue(udIdxOne)
	t = L.GetUserValue(udIdxOne)
	assert.Equal(lua.LUA_TSTRING, t)
	assert.True(L.IsString(-1))
	assert.Equal("onlyuv", L.ToString(-1))
	L.Pop(1)
	L.Pop(2)

	assert.NotNil(L.NewUserDataUv(8, 0))
	assert.Equal(lua.LUA_TUSERDATA, L.Type(-1))
	L.Pop(1)

	assert.NotNil(L.NewUserData(4096))
	assert.Equal(lua.LUA_TUSERDATA, L.Type(-1))
	L.Pop(1)

	L.AtPanic(func(L *lua.State) int {
		err := L.CheckError(lua.LUA_ERRERR)
		assert.Error(err)
		panic(err)
	})
	assert.Panics(func() { L.NewUserData(-1) })
	L.NewUserDataUv(1, -1)

	assert.NotNil(L.NewUserData(12))
	udIdx = L.GetTop()
	mtName := "MyMeta"
	L.NewTable()
	L.SetIMetaTable(udIdx)
	L.NewMetaTable(mtName)
	L.SetIMetaTable(udIdx)

	ptr, err := L.CheckUserData(udIdx, mtName)
	assert.NoError(err)
	assert.True(uintptr(ptr) > 0)

	ptr, err = L.TestUserData(udIdx, mtName)
	assert.NoError(err)
	assert.Equal(ptr, ptr)

	L.NewTable()
	L.SetIMetaTable(udIdx)
	ptr, err = L.TestUserData(udIdx, mtName)
	assert.NoError(err)
	assert.Nil(ptr)

	L.PushInteger(99)
	ptr4, err4 := L.TestUserData(L.GetTop(), mtName)
	assert.NoError(err4)
	assert.Nil(ptr4)

	assert.Panics(func() { L.CheckUserData(L.GetTop(), mtName) })
	assert.Panics(func() { L.CheckUserData(9999, mtName) })
	ptr, err = L.TestUserData(9999, mtName)
	assert.NoError(err)
	assert.Nil(ptr)

	type UserData struct {
		Name string
		Age  int
	}
	var ud *UserData
	ud = (*UserData)(L.NewUserData(int(unsafe.Sizeof(*ud))))
	ud.Name = "John Doe"
	ud.Age = 30

	L.NewTable()
	L.PushString("__index")
	L.PushGoFunction(func(L *lua.State) int {
		userdata := (*UserData)(L.ToUserData(1))
		key := L.ToString(2)
		switch key {
		case "Name":
			L.PushString(userdata.Name)
			return 1
		case "Age":
			L.PushInteger(int64(userdata.Age))
			return 1
		default:
			L.PushNil()
			return 1
		}
	})
	L.SetTable(-3)
	L.SetIMetaTable(-2)

	L.SetGlobal("userDataTest")

	assert.NoError(L.DoFile("testdata/userdata.lua"))
}
