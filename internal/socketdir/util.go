package socketdir

import (
	"net"
	"strings"
)

type Conn interface {
	net.Conn
	CloseWrite() error
}

func replaceFirstLabel(hostname string, replacement string) string {
	dot := strings.IndexByte(hostname, '.')
	if dot == -1 {
		return replacement
	} else {
		return replacement + hostname[dot:]
	}
}
