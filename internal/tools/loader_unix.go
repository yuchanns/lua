//go:build !windows

package tools

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

func BytePtrFromString(s string) (*byte, error) {
	if s == "" {
		return new(byte), nil
	}

	// Check cache first
	cacheMutex.RLock()
	if cached, exists := stringCache[s]; exists {
		cacheMutex.RUnlock()
		return cached, nil
	}
	cacheMutex.RUnlock()

	// Not in cache, create new
	ptr, err := unix.BytePtrFromString(s)
	if err != nil {
		return nil, err
	}

	// Cache the result for common strings (limit cache size to prevent memory leaks)
	cacheMutex.Lock()
	if len(stringCache) < 1000 { // Reasonable cache size limit
		stringCache[s] = ptr
	}
	cacheMutex.Unlock()

	return ptr, nil
}

func BytePtrToString(p *byte) string {
	if p == nil {
		return ""
	}
	return unix.BytePtrToString(p)
}

func LoadLibrary(path string) (uintptr, error) {
	return purego.Dlopen(path, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
}

func FreeLibrary(handle uintptr) error {
	if handle == 0 {
		return nil
	}
	err := purego.Dlclose(handle)
	if err != nil {
		return err
	}
	return nil
}

func GetProcAddress(handle uintptr, name string) (uintptr, error) {
	if handle == 0 {
		return 0, nil
	}
	addr, err := purego.Dlsym(handle, name)
	if err != nil {
		return 0, err
	}
	return addr, nil
}
