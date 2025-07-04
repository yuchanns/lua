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

type funcWithState = func(*Suite, *require.Assertions, *lua.State)

func (s *Suite) testWithState(testFunc funcWithState) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)

		L, err := s.lib.NewState()
		assert.NoError(err)

		t.Cleanup(L.Close)

		testFunc(s, assert, L)
	}
}

type funcWithT = func(*Suite, *require.Assertions, *testing.T)

func (s *Suite) testWithT(testFunc funcWithT) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)

		testFunc(s, assert, t)
	}
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
		if testFunc, ok := method.Func.Interface().(funcWithState); ok {
			t.Run(strings.TrimPrefix(method.Name, "Test"), suite.testWithState(testFunc))
		} else if testFunc, ok := method.Func.Interface().(funcWithT); ok {
			t.Run(strings.TrimPrefix(method.Name, "Test"), suite.testWithT(testFunc))
		}
	}
}
