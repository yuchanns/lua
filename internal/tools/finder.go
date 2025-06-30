package tools

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// FindLocalLuaLibrary searches for Lua library files in the .lua/lib directory
// using wildcards based on the version number and operating system.
// version should be a string like "54" for Lua 5.4
func FindLocalLuaLibrary(version string) (string, error) {
	var pattern string
	
	switch runtime.GOOS {
	case "windows":
		pattern = fmt.Sprintf(".lua/lib/lua*%s*.dll", version)
	case "darwin":
		pattern = fmt.Sprintf(".lua/lib/*lua*%s*.dylib", version)
	default: // linux and other unix-like systems
		pattern = fmt.Sprintf(".lua/lib/*lua*%s*.so*", version)
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

