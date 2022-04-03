package listener // import "src.agwa.name/go-listener"

import (
	"net"
)

// Close every listener in listeners
func CloseAll(listeners []net.Listener) {
	for _, listener := range listeners {
		listener.Close()
	}
}
