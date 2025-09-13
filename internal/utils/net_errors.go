package utils

import (
	"errors"
	"net"
	"net/url"
)

// IsTimeoutOrConnectionError checks whether the provided error is related to a network timeout
// or a connection error.
func IsTimeoutOrConnectionError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() || netErr.Temporary() {
			return true
		}
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if opErr, ok := urlErr.Err.(*net.OpError); ok {
			if _, ok := opErr.Err.(*net.OpError); ok {
				return true
			}
			return true
		}
	}

	return false
}
