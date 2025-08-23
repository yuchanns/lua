//go:build !windows

package lua

import (
	"github.com/ebitengine/purego"
	"golang.org/x/sys/unix"
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
