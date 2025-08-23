package lua_test

import (
	"fmt"

	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/lua"
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

	typ := L.GetField(-1, "greeting")
	assert.Equal(lua.LUA_TSTRING, typ)
	assert.True(L.IsString(-1))
	greeting := L.ToString(-1)
	assert.Equal("hello", greeting)
	L.Pop(1)

	typ = L.GetField(-1, "answer")
	assert.Equal(lua.LUA_TNUMBER, typ)
	assert.True(L.IsInteger(-1))
	answer := L.ToInteger(-1)
	assert.Equal(int64(42), answer)
	L.Pop(1)

	typ = L.GetField(-1, "non_existent")
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

	typ = L.GetField(-1, "nested")
	assert.Equal(lua.LUA_TTABLE, typ)
	assert.Equal(lua.LUA_TTABLE, L.Type(-1))
	typ = L.GetField(-1, "inner_key")
	assert.Equal(lua.LUA_TSTRING, typ)
	assert.True(L.IsString(-1))
	innerValue := L.ToString(-1)
	assert.Equal("inner_value", innerValue)
	L.Pop(2)

	L.PushString("new_greeting")
	L.SetField(-2, "greeting")

	typ = L.GetField(-1, "greeting")
	assert.Equal(lua.LUA_TSTRING, typ)
	newGreeting := L.ToString(-1)
	assert.Equal("new_greeting", newGreeting)
	L.Pop(1)

	L.PushNil()
	L.SetField(-2, "answer")

	typ = L.GetField(-1, "answer")
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
	L.NewTable()
	tableIdx := L.GetTop()

	L.PushString("raw_key")
	L.PushString("raw_value")
	L.RawSet(tableIdx)

	L.PushString("raw_key")
	typ := L.RawGet(tableIdx)
	assert.NotEqual(lua.LUA_TNIL, typ)
	assert.True(L.IsString(-1))
	rawValue := L.ToString(-1)
	assert.Equal("raw_value", rawValue)
	L.Pop(1)

	L.PushString("first_element")
	L.RawSetI(tableIdx, 1)

	L.PushString("second_element")
	L.RawSetI(tableIdx, 2)

	L.PushInteger(100)
	L.RawSetI(tableIdx, 10)

	typ = L.RawGetI(tableIdx, 1)
	assert.NotEqual(lua.LUA_TNIL, typ)
	assert.True(L.IsString(-1))
	first := L.ToString(-1)
	assert.Equal("first_element", first)
	L.Pop(1)

	typ = L.RawGetI(tableIdx, 2)
	assert.NotEqual(lua.LUA_TNIL, typ)
	assert.True(L.IsString(-1))
	second := L.ToString(-1)
	assert.Equal("second_element", second)
	L.Pop(1)

	typ = L.RawGetI(tableIdx, 10)
	assert.NotEqual(lua.LUA_TNIL, typ)
	assert.True(L.IsInteger(-1))
	tenth := L.ToInteger(-1)
	assert.Equal(int64(100), tenth)
	L.Pop(1)

	typ = L.RawGetI(tableIdx, 99)
	assert.Equal(lua.LUA_TNIL, typ)
	assert.True(L.IsNil(-1))
	L.Pop(1)

	testString := "pointer_test"
	testInt := 42

	L.PushString("value_for_string_ptr")
	err := L.RawSetP(tableIdx, &testString)
	assert.NoError(err)

	L.PushInteger(999)
	err = L.RawSetP(tableIdx, &testInt)
	assert.NoError(err)

	typ, err = L.RawGetP(tableIdx, &testString)
	assert.NoError(err)
	assert.NotEqual(lua.LUA_TNIL, typ)
	assert.True(L.IsString(-1))
	ptrValue := L.ToString(-1)
	assert.Equal("value_for_string_ptr", ptrValue)
	L.Pop(1)

	typ, err = L.RawGetP(tableIdx, &testInt)
	assert.NoError(err)
	assert.NotEqual(lua.LUA_TNIL, typ)
	assert.True(L.IsInteger(-1))
	intPtrValue := L.ToInteger(-1)
	assert.Equal(int64(999), intPtrValue)
	L.Pop(1)

	anotherInt := 123
	typ, err = L.RawGetP(tableIdx, &anotherInt)
	assert.NoError(err)
	assert.Equal(lua.LUA_TNIL, typ)
	assert.True(L.IsNil(-1))
	L.Pop(1)

	L.PushString("raw_key")
	L.PushString("new_raw_value")
	L.RawSet(tableIdx)

	L.PushString("raw_key")
	typ = L.RawGet(tableIdx)
	assert.NotEqual(lua.LUA_TNIL, typ)
	newValue := L.ToString(-1)
	assert.Equal("new_raw_value", newValue)
	L.Pop(1)

	L.PushString("raw_key")
	L.PushNil()
	L.RawSet(tableIdx)

	L.PushString("raw_key")
	typ = L.RawGet(tableIdx)
	assert.Equal(lua.LUA_TNIL, typ)
	L.Pop(1)

	L.PushNil()
	L.RawSetI(tableIdx, 1)

	typ = L.RawGetI(tableIdx, 1)
	assert.Equal(lua.LUA_TNIL, typ)
	L.Pop(1)

	L.PushBoolean(true)
	L.RawSetI(tableIdx, 5)

	typ = L.RawGetI(tableIdx, 5)
	assert.Equal(lua.LUA_TBOOLEAN, typ)
	assert.True(L.ToBoolean(-1))
	L.Pop(1)

	L.NewTable()
	L.PushString("nested_value")
	L.RawSetI(-2, 1)
	L.RawSetI(tableIdx, 20)

	typ = L.RawGetI(tableIdx, 20)
	assert.Equal(lua.LUA_TTABLE, typ)
	typ = L.RawGetI(-1, 1)
	assert.Equal(lua.LUA_TSTRING, typ)
	nestedValue := L.ToString(-1)
	assert.Equal("nested_value", nestedValue)
	L.Pop(2)

	L.Pop(1)

	assert.Equal(0, L.GetTop())
}

func (s *Suite) TestTableNext(assert *require.Assertions, L *lua.State) {
	L.NewTable()
	tableIdx := L.GetTop()

	L.NewTable()
	emptyTableIdx := L.GetTop()

	L.PushNil()
	assert.False(L.Next(emptyTableIdx))

	L.PushString("string_key")
	L.PushString("string_value")
	L.RawSet(tableIdx)

	L.PushInteger(42)
	L.RawSetI(tableIdx, 1)

	L.PushBoolean(true)
	L.RawSetI(tableIdx, 5)

	L.PushNumber(3.14)
	L.RawSetI(tableIdx, 10)

	L.PushString("nested_table")
	L.NewTable()
	L.PushString("nested")
	L.RawSetI(-2, 1)
	L.RawSet(tableIdx)

	expectedKeys := make(map[string]bool)
	expectedKeys["string_key"] = false
	expectedKeys["1"] = false
	expectedKeys["5"] = false
	expectedKeys["10"] = false
	expectedKeys["nested_table"] = false

	keyCount := 0
	L.PushNil()
	for L.Next(tableIdx) {
		keyCount++

		keyType := L.Type(-2)
		valueType := L.Type(-1)

		var keyStr string
		switch keyType {
		case lua.LUA_TSTRING:
			keyStr = L.ToString(-2)
		case lua.LUA_TNUMBER:
			keyStr = fmt.Sprintf("%d", L.ToInteger(-2))
		}

		switch keyStr {
		case "string_key":
			assert.Equal(lua.LUA_TSTRING, valueType)
			assert.Equal("string_value", L.ToString(-1))
			expectedKeys["string_key"] = true
		case "1":
			assert.Equal(lua.LUA_TNUMBER, valueType)
			assert.Equal(int64(42), L.ToInteger(-1))
			expectedKeys["1"] = true
		case "5":
			assert.Equal(lua.LUA_TBOOLEAN, valueType)
			assert.True(L.ToBoolean(-1))
			expectedKeys["5"] = true
		case "10":
			assert.Equal(lua.LUA_TNUMBER, valueType)
			assert.Equal(3.14, L.ToNumber(-1))
			expectedKeys["10"] = true
		case "nested_table":
			assert.Equal(lua.LUA_TTABLE, valueType)
			expectedKeys["nested_table"] = true
		}

		L.Pop(1)
	}

	assert.Equal(5, keyCount)
	for key, found := range expectedKeys {
		assert.True(found, "Key %s not found during iteration", key)
	}

	L.PushString("iteration_key")
	L.PushString("iteration_value")
	L.RawSet(tableIdx)

	newKeyCount := 0
	L.PushNil()
	for L.Next(tableIdx) {
		newKeyCount++
		L.Pop(1)
	}
	assert.Equal(6, newKeyCount)

	L.PushString("iteration_key")
	L.PushNil()
	L.RawSet(tableIdx)

	finalKeyCount := 0
	L.PushNil()
	for L.Next(tableIdx) {
		finalKeyCount++
		L.Pop(1)
	}
	assert.Equal(5, finalKeyCount)

	L.NewTable()
	sparseTableIdx := L.GetTop()

	L.PushString("value1")
	L.RawSetI(sparseTableIdx, 1)

	L.PushString("value100")
	L.RawSetI(sparseTableIdx, 100)

	L.PushString("value1000")
	L.RawSetI(sparseTableIdx, 1000)

	sparseCount := 0
	indices := make([]int64, 0)
	L.PushNil()
	for L.Next(sparseTableIdx) {
		sparseCount++
		key := L.ToInteger(-2)
		indices = append(indices, key)
		L.Pop(1)
	}

	assert.Equal(3, sparseCount)
	assert.Contains(indices, int64(1))
	assert.Contains(indices, int64(100))
	assert.Contains(indices, int64(1000))

	L.NewTable()
	mixedTableIdx := L.GetTop()

	L.PushString("array_val1")
	L.RawSetI(mixedTableIdx, 1)

	L.PushString("array_val2")
	L.RawSetI(mixedTableIdx, 2)

	L.PushString("hash_value")
	L.SetField(mixedTableIdx, "hash_key")

	L.PushNumber(99.9)
	L.SetField(mixedTableIdx, "number_key")

	mixedCount := 0
	arrayKeys := 0
	hashKeys := 0

	L.PushNil()
	for L.Next(mixedTableIdx) {
		mixedCount++

		if L.Type(-2) == lua.LUA_TNUMBER {
			num := L.ToInteger(-2)
			if num == 1 || num == 2 {
				arrayKeys++
			}
		} else if L.Type(-2) == lua.LUA_TSTRING {
			hashKeys++
		}

		L.Pop(1)
	}

	assert.Equal(4, mixedCount)
	assert.Equal(2, arrayKeys)
	assert.Equal(2, hashKeys)

	L.Pop(4)
	assert.Equal(0, L.GetTop())
}

func (s *Suite) TestTableMeta(assert *require.Assertions, L *lua.State) {
	L.NewTable()
	tblIdx := L.GetTop()
	L.NewTable()
	mtIdx := L.GetTop()

	L.PushString("__call")
	L.PushGoFunction(func(L *lua.State) int {
		L.PushString("called")
		return 1
	})
	L.SetTable(mtIdx)

	L.SetIMetaTable(tblIdx)
	L.PushValue(tblIdx)
	assert.Equal(1, L.GeIMetaTable(-1))
	assert.Equal(lua.LUA_TTABLE, L.Type(-1))
	L.Pop(2)

	assert.Equal(1, L.GeIMetaTable(tblIdx))
	typ := L.GetMetaField(tblIdx, "__call")
	assert.Equal(lua.LUA_TFUNCTION, typ)

	has := L.CallMeta(tblIdx, "__call")
	assert.True(has)
	assert.Equal(lua.LUA_TSTRING, L.Type(-1))
	assert.Equal("called", L.ToString(-1))

	L.SetTop(tblIdx)

	has = L.NewMetaTable("mt")
	assert.False(has)
	L.Pop(1)

	has = L.NewMetaTable("mt")
	assert.True(has)
	L.PushString("__call")
	L.PushGoFunction(func(L *lua.State) int {
		L.PushString("mt called")
		return 1
	})
	L.SetTable(-3)
	L.SetTop(tblIdx)

	L.SetMetaTable("mt")
	typ = L.GetMetaField(tblIdx, "__call")
	assert.Equal(lua.LUA_TFUNCTION, typ)

	typ = L.GetMetaTable("mt")
	assert.Equal(lua.LUA_TTABLE, typ)
	typ = L.GetMetaTable("non_existent")
	assert.Equal(lua.LUA_TNIL, typ)

	has = L.CallMeta(tblIdx, "__call")
	assert.True(has)
	assert.Equal(lua.LUA_TSTRING, L.Type(-1))
	assert.Equal("mt called", L.ToString(-1))
}
