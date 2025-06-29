package lua_test

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/lua"
)

type Suite struct {
	state *lua.State
}

func (s *Suite) Setup() (err error) {
	var path string

	// INFO: Currently, we only support lua5.4.8.
	// More versions can be added later.
	switch runtime.GOOS {
	case "darwin":
		path = "/opt/homebrew/Cellar/lua/5.4.8/lib/liblua.dylib"
	case "linux":
		path = "/lib64/liblua-5.4.so"
	}

	s.state, err = lua.NewState(path)

	if err != nil {
		return
	}

	return
}

func (s *Suite) TearDown() {
	if s.state != nil {
		s.state.Close()
	}
}

func TestSuite(t *testing.T) {
	assert := require.New(t)

	suite := &Suite{}

	assert.NoError(suite.Setup())

	t.Cleanup(suite.TearDown)

	tt := reflect.TypeOf(suite)
	for i := range tt.NumMethod() {
		method := tt.Method(i)
		testFunc, ok := method.Func.Interface().(func(*Suite, *require.Assertions))
		if !ok {
			continue
		}
		t.Run(strings.TrimLeft(method.Name, "Test"), func(t *testing.T) {
			testFunc(suite, require.New(t))
		})
	}
}

func (s *Suite) TestBasic(assert *require.Assertions) {
	state := s.state
	state.PushGoFunction(func(x float64) float64 {
		return x * 2
	})
	assert.NoError(state.SetGlobal("double_number"))
	assert.NoError(state.DoString(`print(double_number(21))`))
}
