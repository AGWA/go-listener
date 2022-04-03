package proxy

import (
	"net"
)

type proxyConn struct {
	net.Conn
	localAddr  net.Addr
	remoteAddr net.Addr
}

func (conn *proxyConn) LocalAddr() net.Addr {
	return conn.localAddr
}

func (conn *proxyConn) RemoteAddr() net.Addr {
	return conn.remoteAddr
}
