//go:build windows

package tools

import (
	"golang.org/x/sys/windows"
	"sync"
)

// String cache to avoid repeated allocations for common strings
var (
	stringCache = make(map[string]*byte)
	cacheMutex  sync.RWMutex
)

func BytePtrFromString(s string) (*byte, error) {
	if s == "" {
		return new(byte), nil
	}

	ptr, err := windows.BytePtrFromString(s)
	if err != nil {
		return nil, err
	}

	return ptr, nil
}

func BytePtrToString(p *byte) string {
	if p == nil {
		return ""
	}
	return windows.BytePtrToString(p)
}

func LoadLibrary(path string) (uintptr, error) {
	handle, err := windows.LoadLibrary(path)
	if err != nil {
		return 0, err
	}
	return uintptr(handle), nil
}

func FreeLibrary(handle uintptr) error {
	if handle == 0 {
		return nil
	}
	err := windows.FreeLibrary(windows.Handle(handle))
	if err != nil {
		return err
	}
	return nil
}

func GetProcAddress(handle uintptr, name string) (uintptr, error) {
	if handle == 0 {
		return 0, nil
	}
	proc, err := windows.GetProcAddress(windows.Handle(handle), name)
	if err != nil {
		return 0, err
	}
	return proc, nil
}
