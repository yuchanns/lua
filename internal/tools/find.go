package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// FindLuaLibrary performs a global search for Lua dynamic library
func FindLuaLibrary(version string) (string, error) {
	var searchPatterns []string
	var searchDirs []string

	switch runtime.GOOS {
	case "darwin":
		if version != "" {
			// Version-specific patterns for macOS
			searchPatterns = []string{
				fmt.Sprintf("liblua%s*.dylib", version),
				fmt.Sprintf("liblua-%s*.dylib", version),
				fmt.Sprintf("liblua.%s*.dylib", version),
			}
		} else {
			// General pattern for macOS
			searchPatterns = []string{"liblua*.dylib"}
		}
		searchDirs = []string{"/usr", "/opt", "/lib", "/System"}
	case "linux":
		if version != "" {
			// Version-specific patterns for Linux
			searchPatterns = []string{
				fmt.Sprintf("liblua%s*.so*", version),
				fmt.Sprintf("liblua-%s*.so*", version),
				fmt.Sprintf("liblua.%s*.so*", version),
			}
		} else {
			// General pattern for Linux
			searchPatterns = []string{"liblua*.so*"}
		}
		searchDirs = []string{"/usr", "/lib", "/lib64", "/opt"}
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Try each pattern in order of preference
	for _, pattern := range searchPatterns {
		// Try using find command for efficient global search
		for _, dir := range searchDirs {
			if path, err := findUsingCommand(dir, pattern); err == nil && path != "" {
				return path, nil
			}
		}

		// Fallback to Go's filepath.Walk if find command fails
		for _, dir := range searchDirs {
			if path, err := findUsingWalk(dir, pattern); err == nil && path != "" {
				return path, nil
			}
		}
	}

	if version != "" {
		return "", fmt.Errorf("lua library version %s not found on this system", version)
	}
	return "", fmt.Errorf("lua library not found on this system")
}

// findUsingCommand uses system find command for efficient search
func findUsingCommand(searchDir, pattern string) (string, error) {
	if _, err := os.Stat(searchDir); os.IsNotExist(err) {
		return "", err
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd = exec.Command("find", searchDir, "-name", pattern, "-type", "f", "-print", "-quit")
	default:
		return "", fmt.Errorf("find command not supported on %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	path := strings.TrimSpace(string(output))
	if path != "" && isValidLibrary(path) {
		return path, nil
	}

	return "", fmt.Errorf("no valid library found")
}

// findUsingWalk uses Go's filepath.Walk for search (fallback method)
func findUsingWalk(searchDir, pattern string) (string, error) {
	if _, err := os.Stat(searchDir); os.IsNotExist(err) {
		return "", err
	}

	var foundPath string
	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Continue walking even if we can't access some directories
			return nil
		}

		if info.IsDir() {
			return nil
		}

		matched, err := filepath.Match(pattern, info.Name())
		if err != nil {
			return nil
		}

		if matched && isValidLibrary(path) {
			foundPath = path
			return fmt.Errorf("found") // Stop walking
		}

		return nil
	})

	if foundPath != "" {
		return foundPath, nil
	}

	if err != nil && err.Error() == "found" {
		return foundPath, nil
	}

	return "", fmt.Errorf("library not found in %s", searchDir)
}

// isValidLibrary checks if the found file is a valid library
func isValidLibrary(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if it's a regular file and not too small
	if !info.Mode().IsRegular() || info.Size() < 1024 {
		return false
	}

	return true
}
