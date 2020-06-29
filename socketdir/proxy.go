package socketdir

import (
	"io"

	"src.agwa.name/go-listener/proxy"
)

type ProxyProto string

const (
	ProtocolPlain ProxyProto = "plain"
	ProtocolProxy ProxyProto = "proxy"
)

var proxyProtos = []ProxyProto{ProtocolPlain, ProtocolProxy}

func ProxyConnection(protocol ProxyProto, clientConn Conn, backendConn Conn) {
	switch protocol {
	case ProtocolPlain:
		proxyConnectionPlain(clientConn, backendConn)
	case ProtocolProxy:
		proxyConnectionProxy(clientConn, backendConn)
	default:
		panic("ProxyConnection: illegal proxy protocol value")
	}
}

func proxyConnectionPlain(clientConn Conn, backendConn Conn) {
	go func() {
		io.Copy(backendConn, clientConn)
		backendConn.CloseWrite()
	}()

	io.Copy(clientConn, backendConn)
}

func proxyConnectionProxy(clientConn Conn, backendConn Conn) {
	// TODO: if clientConn is a *tls.Conn, consider also sending hostname and protocol using PP2_TYPE_AUTHORITY and PP2_TYPE_ALPN... we could also send detailed TLS info using PP2_TYPE_SSL
	header := proxy.Header{RemoteAddr: clientConn.RemoteAddr(), LocalAddr: clientConn.LocalAddr()}
	if _, err := backendConn.Write(header.Format()); err != nil {
		return
	}

	proxyConnectionPlain(clientConn, backendConn)
}
