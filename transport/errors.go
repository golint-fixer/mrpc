package transport

import "errors"

// ErrTimeout is returned when the request timeouts
var ErrTimeout = errors.New("Transport: timeout")

// IsTimeout checks if the error is a timeout
func IsTimeout(err error) bool {
	return err == ErrTimeout
}
