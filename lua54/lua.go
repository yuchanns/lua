package lua

import (
	"fmt"

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

func (l *Lib) NewState() (state *State, err error) {
	if l.ffi == nil {
		return nil, fmt.Errorf("Lua library is closed")
	}

	state = newState(l.ffi)

	return
}
