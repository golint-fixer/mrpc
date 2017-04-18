package transport

import "errors"

var ErrTimeout = errors.New("Transport: timeout")

func IsTimeout(err error) bool {
	return err == ErrTimeout
}
