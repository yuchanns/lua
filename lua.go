package lua

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Lib represents a loaded Lua 5.4 dynamic library binding in Go. It provides access to library-level operations and state creation (see: https://www.lua.org/manual/5.4/manual.html#4.3).
type Lib struct {
	ffi *ffi
}

// New loads a Lua 5.4 dynamic library from the given path and returns a Lib for further state management.
// Returns an error if the library cannot be loaded.
func New(path string) (lib *Lib, err error) {
	ffi, err := newFFI(path)
	if err != nil {
		return
	}

	lib = &Lib{
		ffi: ffi,
	}

	return
}

// Close releases the loaded Lua dynamic library and any resources associated with it in this Lib instance.
func (l *Lib) Close() (err error) {
	if l.ffi == nil {
		return
	}

	err = freeLibrary(l.ffi.lib)
	if err == nil {
		l.ffi = nil
	}
	return
}

// NewState creates a new Lua runtime state from this Lib (binding to the dynamic library).
// Additional options may be provided for custom allocators and user data.
// Returns a State and possibly an error if the library is closed.
func (l *Lib) NewState(o ...stateOptFunc) (state *State, err error) {
	if l.ffi == nil {
		return nil, fmt.Errorf("lua library is closed")
	}

	opt := &stateOpt{}
	for _, fn := range o {
		fn(opt)
	}

	state = newState(l.ffi, opt)

	return
}

// FFI returns the underlying ffi instance for advanced usage.
func (l *Lib) FFI() *ffi {
	return l.ffi
}

// stateOptFunc is an option setter for customizing State creation (internal use).
type stateOptFunc func(o *stateOpt)

// WithAlloc sets a custom memory allocation function for the Lua state.
// SAFETY: It is guaranteed that the user data pointer will be valid
// during the lifetime of the Lua state.
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
