package listeners // import "src.agwa.name/go-listeners"

import (
	"net"
)

func CloseListeners(listeners []net.Listener) {
	for _, listener := range listeners {
		listener.Close()
	}
}
