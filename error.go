package lua

import "fmt"

// Error represents a Lua error with its status code and corresponding message.
// It is returned by many operations when faults occur, matching the error codes of the Lua C API.
// See: https://www.lua.org/manual/5.4/manual.html#4.4
type Error struct {
	status  int
	message string
}

// Error implements the error interface for Lua Error, returning a formatted error string.
func (e *Error) Error() string {
	return fmt.Sprintf("Lua Error %d: %s", e.status, e.message)
}

// Status returns the status code associated with the Lua error.
func (e *Error) Status() int {
	return e.status
}

// Message returns the string message of the Lua error.
func (e *Error) Message() string {
	return e.message
}

// UnprotectedError represents an error that occurs when an operation is attempted on a Lua state
// that called without pcallk or pcall.
type UnprotectedError struct {
	message string
}

func (e *UnprotectedError) Error() string {
	return fmt.Sprintf("Unprotected Error in call to Lua API (%s)", e.message)
}
