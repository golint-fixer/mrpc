package transport

import "errors"

var errTimeout = errors.New("Transport: timeout")

func IsTimeout(err error) bool {
	return err == errTimeout
}
