package lua_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/lua"
)

type Suite struct {
}

func findLocalLuaLibrary(version string) (string, error) {
	var pattern string

	switch runtime.GOOS {
	case "windows":
		pattern = fmt.Sprintf("lua%s/.lua/lib/lua*%s*.dll", version, version)
	case "darwin":
		pattern = fmt.Sprintf("lua%s/.lua/lib/*lua*%s*.dylib", version, version)
	default: // linux and other unix-like systems
		pattern = fmt.Sprintf("lua%s/.lua/lib/*lua*%s*.so*", version, version)
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to search for Lua library with pattern %s: %w", pattern, err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no Lua library found matching pattern %s", pattern)
	}

	// Return the first match if multiple files are found
	return matches[0], nil
}

func (s *Suite) Setup() (err error) {
	version := os.Getenv("LUA_VERSION")
	if version == "" {
		version = "54"
	}
	path, err := findLocalLuaLibrary(version)
	if err != nil {
		return
	}
	err = lua.Init(path)
	if err != nil {
		return
	}

	return nil
}

func (s *Suite) TearDown() {
	_ = lua.Deinit()
}

type funcWithState = func(*Suite, *require.Assertions, *lua.State)

func (s *Suite) testWithState(testFunc funcWithState) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)

		L := lua.NewState()

		L.OpenLibs()

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

	L := lua.NewState()
	t.Cleanup(L.Close)

	t.Logf("Running tests for lua version %v", L.Version())

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
