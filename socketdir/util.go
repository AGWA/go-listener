package socketdir

import (
	"net"
	"os"
	"strings"
)

type Conn interface {
	net.Conn
	CloseWrite() error
}

func directoryExists(dir string) bool {
	info, err := os.Stat(dir)
	if err == nil && info.IsDir() {
		return true
	}
	if err == nil {
		//log.Printf("Ignoring %s because it is not a directory", dir)
	} else if !os.IsNotExist(err) {
		//log.Printf("Ignoring %s due to stat error: %s", dir, err)
	}
	return false
}

func replaceFirstLabel(hostname string, replacement string) string {
	dot := strings.IndexByte(hostname, '.')
	if dot == -1 {
		return replacement
	} else {
		return replacement + hostname[dot:]
	}
}

func canonicalizeHostname(hostname string) string {
	hostname = strings.ToLower(hostname)
	hostname = strings.TrimRight(hostname, ".")
	return hostname
}
