package lua_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/lua/internal/tools"
	"go.yuchanns.xyz/lua/lua54"
)

type Suite struct {
	lib *lua.Lib
}

func (s *Suite) Setup() (err error) {
	path, err := tools.FindLocalLuaLibrary("54")
	if err != nil {
		return
	}
	s.lib, err = lua.New(path)
	if err != nil {
		return
	}

	return nil
}

func (s *Suite) TearDown() {
	if s.lib == nil {
		return
	}
	s.lib.Close()
}

func TestSuite(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	suite := &Suite{}

	assert.NoError(suite.Setup())

	t.Cleanup(suite.TearDown)

	tt := reflect.TypeOf(suite)
	for i := range tt.NumMethod() {
		method := tt.Method(i)
		testFunc, ok := method.Func.Interface().(func(*Suite, *require.Assertions, *lua.State))
		if !ok {
			continue
		}

		t.Run(strings.TrimLeft(method.Name, "Test"), func(t *testing.T) {
			t.Parallel()

			assert := require.New(t)

			L, err := suite.lib.NewState()
			assert.NoError(err)

			t.Cleanup(L.Close)

			testFunc(suite, assert, L)
		})
	}
}

func (s *Suite) TestBasic(assert *require.Assertions, L *lua.State) {
	L.PushGoFunction(func(x float64) float64 {
		return x * 2
	})
	assert.NoError(L.SetGlobal("double_number"))
	assert.NoError(L.DoString(`print(double_number(21))`))
}
