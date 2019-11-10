package listeners // import "src.agwa.name/go-listeners"

import (
	"net"
)

func CloseTCP(listeners []*net.TCPListener) {
	for _, listener := range listeners {
		listener.Close()
	}
}

func CloseUDP(listeners []*net.UDPConn) {
	for _, listener := range listeners {
		listener.Close()
	}
}
