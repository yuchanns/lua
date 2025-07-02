package lua

import (
	"fmt"
	"unsafe"

	"go.yuchanns.xyz/lua/internal/tools"
)

type Lib struct {
	ffi *ffi
}

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

func (l *Lib) Close() {
	if l.ffi == nil {
		return
	}

	defer tools.FreeLibrary(l.ffi.lib)

	l.ffi = nil
}

func (l *Lib) NewState(o ...stateOptFunc) (state *State, err error) {
	if l.ffi == nil {
		return nil, fmt.Errorf("Lua library is closed")
	}

	var opt *stateOpt
	if len(o) > 0 {
		opt = &stateOpt{}
	}
	for _, fn := range o {
		fn(opt)
	}

	state = newState(l.ffi, opt)

	return
}

type stateOptFunc func(o *stateOpt)

func WithAlloc[T any](
	fn func(ud *T, ptr unsafe.Pointer, osize, nsize int) unsafe.Pointer,
	ud *T,
) stateOptFunc {
	return func(o *stateOpt) {
		o.alloc = func(ud, ptr unsafe.Pointer, osize, nsize int) unsafe.Pointer {
			t := (*T)(ud)
			return fn(t, ptr, osize, nsize)
		}
		o.userData = unsafe.Pointer(ud)
	}
}
