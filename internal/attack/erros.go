package attack

import "fmt"

type InvalidStatusCode struct {
	Code int
}

func (e *InvalidStatusCode) Error() string {
	return fmt.Sprintf("non-success status code (200–399): %d", e.Code)
}
