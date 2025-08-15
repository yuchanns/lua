//go:build !windows

package lua

import (
	"github.com/ebitengine/purego"
	"golang.org/x/sys/unix"
	"sync"
)

// String cache to avoid repeated allocations for common strings
var (
	stringCache = make(map[string]*byte)
	cacheMutex  sync.RWMutex
)

func bytePtrFromString(s string) (*byte, error) {
	if s == "" {
		return new(byte), nil
	}

	ptr, err := unix.BytePtrFromString(s)
	if err != nil {
		return nil, err
	}

	return ptr, nil
}

func bytePtrToString(p *byte) string {
	if p == nil {
		return ""
	}
	return unix.BytePtrToString(p)
}

func loadLibrary(path string) (uintptr, error) {
	return purego.Dlopen(path, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
}

func freeLibrary(handle uintptr) error {
	if handle == 0 {
		return nil
	}
	err := purego.Dlclose(handle)
	if err != nil {
		return err
	}
	return nil
}

func getProcAddress(handle uintptr, name string) (uintptr, error) {
	if handle == 0 {
		return 0, nil
	}
	addr, err := purego.Dlsym(handle, name)
	if err != nil {
		return 0, err
	}
	return addr, nil
}
