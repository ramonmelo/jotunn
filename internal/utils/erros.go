package utils

import "fmt"

// CSRFFieldError represents an error related to CSRF field validation.
type CSRFFieldError struct {
	Message string
	Code    int
}

// Error implements the error interface for CSRFFieldError, returning a formatted error message
// that includes the error code and message.
func (e *CSRFFieldError) Error() string {
	return fmt.Sprintf("Erro %d: %s", e.Code, e.Message)
}
