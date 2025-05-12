package utils

import "fmt"

type CSRFFieldError struct {
	Message string
	Code    int
}

func (e *CSRFFieldError) Error() string {
	return fmt.Sprintf("Erro %d: %s", e.Code, e.Message)
}
