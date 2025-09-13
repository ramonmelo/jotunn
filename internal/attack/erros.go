package attack

import "fmt"

// InvalidStatusCode represents an error for non-success HTTP status codes.
type InvalidStatusCode struct {
	Code int
}

// Error implements the error interface for InvalidStatusCode, returning a formatted error message
func (e *InvalidStatusCode) Error() string {
	return fmt.Sprintf("non-success status code (200–399): %d", e.Code)
}
