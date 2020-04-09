package proxy

import (
	"net"
)

type Conn struct {
	net.Conn
	localAddr  net.Addr
	remoteAddr net.Addr
}

func (conn *Conn) LocalAddr() net.Addr {
	return conn.localAddr
}

func (conn *Conn) RemoteAddr() net.Addr {
	return conn.remoteAddr
}
