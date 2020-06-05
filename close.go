package listener // import "src.agwa.name/go-listener"

import (
	"net"
)

func CloseAll(listeners []net.Listener) {
	for _, listener := range listeners {
		listener.Close()
	}
}
