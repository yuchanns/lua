package lua

import (
	"fmt"
	"reflect"
	"unsafe"

	"go.yuchanns.xyz/lua/internal/tools"
)

// PushGoFunction registers a Go function as a Lua C function.
// Panics if the f is not a function or if it has unsupported argument/return types.
func (s *State) PushGoFunction(f any) {
	// Compile function metadata once during registration
	metadata := compileFuncMetadata(f, s.ffi)

	// Create optimized LuaCFunction using pre-compiled metadata
	var fn LuaCFunction = func(L unsafe.Pointer) int {
		// Fast argument conversion using pre-compiled converters
		args := make([]reflect.Value, metadata.numArgs)
		for i, converter := range metadata.argConverters {
			args[i] = converter(s, L, i+1)
		}

		// Call function with converted arguments
		results := metadata.fn.Call(args)

		// Fast result pushing using pre-compiled pushers
		for i, pusher := range metadata.resultPushers {
			pusher(s, L, results[i])
		}

		return metadata.numResults
	}

	s.PushCFunction(fn)
}

// argConverter converts Lua stack value to Go reflect.Value
type argConverter func(*State, unsafe.Pointer, int) reflect.Value

// resultPusher pushes Go reflect.Value to Lua stack
type resultPusher func(*State, unsafe.Pointer, reflect.Value)

// funcMetadata contains pre-compiled function call metadata
type funcMetadata struct {
	argConverters []argConverter
	resultPushers []resultPusher
	fn            reflect.Value
	numArgs       int
	numResults    int
}

// createArgConverter creates a type-specific argument converter
func createArgConverter(argType reflect.Type, ffi *ffi) argConverter {
	switch argType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(s *State, L unsafe.Pointer, idx int) reflect.Value {
			val := ffi.LuaLCheckinteger(L, idx)
			if argType.Kind() == reflect.Int {
				return reflect.ValueOf(int(val))
			}
			return reflect.ValueOf(val).Convert(argType)
		}
	case reflect.Float32, reflect.Float64:
		return func(s *State, L unsafe.Pointer, idx int) reflect.Value {
			val := ffi.LuaLChecknumber(L, idx)
			if argType.Kind() == reflect.Float32 {
				return reflect.ValueOf(float32(val))
			}
			return reflect.ValueOf(val)
		}
	case reflect.String:
		return func(s *State, L unsafe.Pointer, idx int) reflect.Value {
			p := ffi.LuaLChecklstring(L, idx, nil)
			return reflect.ValueOf(tools.BytePtrToString(p))
		}
	case reflect.Bool:
		return func(s *State, L unsafe.Pointer, idx int) reflect.Value {
			ffi.LuaLChecktype(L, idx, LUA_TBOOLEAN)
			return reflect.ValueOf(ffi.LuaToboolean(L, idx) != 0)
		}
	default:
		panic(fmt.Sprintf("unsupported argument type: %s", argType.Kind()))
	}
}

// createResultPusher creates a type-specific result pusher
func createResultPusher(resultType reflect.Type, ffi *ffi) resultPusher {
	switch resultType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(s *State, L unsafe.Pointer, val reflect.Value) {
			ffi.LuaPushinteger(L, val.Int())
		}
	case reflect.Float32, reflect.Float64:
		return func(s *State, L unsafe.Pointer, val reflect.Value) {
			ffi.LuaPushnumber(L, val.Float())
		}
	case reflect.String:
		return func(s *State, L unsafe.Pointer, val reflect.Value) {
			str := val.String()
			nptr, err := tools.BytePtrFromString(str)
			if err != nil {
				panic(fmt.Sprintf("failed to convert string to byte pointer: %v", err))
			}
			ffi.LuaPushlstring(L, nptr, len(str))
		}
	case reflect.Bool:
		return func(s *State, L unsafe.Pointer, val reflect.Value) {
			var b int
			if val.Bool() {
				b = 1
			}
			ffi.LuaPushboolean(L, b)
		}
	default:
		panic(fmt.Sprintf("unsupported return type: %s", resultType.Kind()))
	}
}

// compileFuncMetadata performs one-time reflection analysis
func compileFuncMetadata(f any, ffi *ffi) *funcMetadata {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("expected a function, got %s", t.Kind()))
	}
	v := reflect.ValueOf(f)

	numArgs := t.NumIn()
	numResults := t.NumOut()

	metadata := &funcMetadata{
		argConverters: make([]argConverter, numArgs),
		resultPushers: make([]resultPusher, numResults),
		fn:            v,
		numArgs:       numArgs,
		numResults:    numResults,
	}

	// Pre-compile argument converters
	for i := 0; i < numArgs; i++ {
		metadata.argConverters[i] = createArgConverter(t.In(i), ffi)
	}

	// Pre-compile result pushers
	for i := 0; i < numResults; i++ {
		metadata.resultPushers[i] = createResultPusher(t.Out(i), ffi)
	}

	return metadata
}
