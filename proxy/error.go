package proxy

import (
	"net"
)

type acceptError struct {
	error
	temporary bool
}

func (err *acceptError) Temporary() bool {
	return err.temporary
}

func (err *acceptError) Timeout() bool {
	return false
}

var _ net.Error = (*acceptError)(nil) // Cause compile error if acceptError does not implement net.Error interface
