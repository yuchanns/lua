package lua

import "fmt"

// Error represents a Lua error with status code and error message.
type Error struct {
	status  int
	message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Lua Error %d: %s", e.status, e.message)
}

func (e *Error) Status() int {
	return e.status
}

func (e *Error) Message() string {
	return e.message
}
