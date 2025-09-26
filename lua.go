package lua

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
)

// luaLib is a singleton instance for managing the loaded Lua dynamic library.
// Use Init to load a library before using it.
// Use Deinit to release the library when it is no longer needed.
var luaLib = new(lib)

type lib struct {
	ffi *ffi
}

func (l *lib) assert() {
	if l == nil || l.ffi == nil {
		panic("lua library is not loaded, call lua.Init to load a library first")
	}
}

// Init loads a Lua dynamic library from the given path to the global singleton instance.
// Returns an error if the library cannot be loaded.
// Calling Init for multiple times without deinit the previous library will result in an error.
func Init(path string) (err error) {
	if luaLib != nil && luaLib.ffi != nil {
		return fmt.Errorf("previous lua library is not closed, call lua.Deinit first")
	}

	ffi, err := newFFI(path)
	if err != nil {
		return
	}

	*luaLib = lib{
		ffi: ffi,
	}

	return
}

// Deinit releases the loaded Lua dynamic library from the global singleton instance.
// Panics if the library is not initialized.
func Deinit() (err error) {
	luaLib.assert()

	err = freeLibrary(luaLib.ffi.lib)
	if err == nil {
		luaLib.ffi = nil
	}
	return
}

// NewState creates a new Lua runtime state.
// Additional options may be provided for custom allocators and user data.
// Returns a State
// Panics if the library is not initialized.
func NewState(o ...stateOptFunc) (state *State) {
	luaLib.assert()

	opt := &stateOpt{}
	for _, fn := range o {
		fn(opt)
	}

	state = newState(opt)

	return
}

// BuildState create a existing Lua state from a given lua_State pointer.
// Panics if the library is not initialized.
func BuildState(L unsafe.Pointer, o ...stateOptFunc) (state *State) {
	luaLib.assert()

	opt := &stateOpt{}
	for _, fn := range o {
		fn(opt)
	}

	state = &State{
		ffi:  luaLib.ffi,
		luaL: L,

		unwindingProtection: !opt.withoutUwindingProtection,
	}

	return
}

// FFI returns the underlying ffi instance for advanced usage.
// Panics if the library is not initialized.
func FFI() *ffi {
	luaLib.assert()

	return luaLib.ffi
}

// NewCallback creates a C function pointer that wraps a Go function
// that accepts a State and returns an int.
// The returned pointer can be used with PushCFunction or PushCClousure.
// Due to the limitation of Purego, only a limited number (2000) of callbacks
// may be created in a single Go process, and any memory allocated for
// these callbacks is never released.
func NewCallback(f GoFunc, o ...stateOptFunc) uintptr {
	return purego.NewCallback(func(L unsafe.Pointer) int {
		state := BuildState(L, o...)
		return f(state)
	})
}

// stateOptFunc is an option setter for customizing State creation (internal use).
type stateOptFunc func(o *stateOpt)

// WithAlloc sets a custom memory allocation function for the Lua state.
// Due to the limitation of Purego, only a limited number of callbacks may be created in a single Go
// process, and any memory allocated for these callbacks is never released.
// UNSAFE: The userdata must be a pointer type, and it is the caller's responsibility to ensure
// that the pointer remains valid for the lifetime of the Lua state.
func WithAlloc[T any](
	fn func(ud *T, ptr unsafe.Pointer, osize, nsize int) unsafe.Pointer,
	ud *T,
) stateOptFunc {
	return func(o *stateOpt) {
		o.alloc = purego.NewCallback(func(ud, ptr unsafe.Pointer, osize, nsize int) unsafe.Pointer {
			t := (*T)(ud)
			return fn(t, ptr, osize, nsize)
		})
		o.userData = unsafe.Pointer(ud)
	}
}

// WithoutUnwindingProtection indicates whether this state is created without goroutine stack
// unwinding protected mode. Lua use `setjmp/longjmp` to handle errors, which is not
// compatible with Go's goroutine stack unwinding and will cause syscall frames no longer
// available on callback to Go code.
func WithoutUnwindingProtection() stateOptFunc {
	return func(o *stateOpt) {
		o.withoutUwindingProtection = true
	}
}
