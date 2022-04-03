package proxy

import (
	"errors"
	"fmt"
	"net"
	"time"
)

type Listener struct {
	inner  net.Listener
	conns  chan net.Conn
	errors chan error
	done   chan struct{}
}

func NewListener(inner net.Listener) *Listener {
	listener := &Listener{
		inner:  inner,
		conns:  make(chan net.Conn),
		errors: make(chan error),
		done:   make(chan struct{}),
	}
	go listener.handleAccepts()
	return listener
}

func (listener *Listener) Accept() (net.Conn, error) {
	select {
	case conn := <-listener.conns:
		return conn, nil
	case err := <-listener.errors:
		return nil, err
	case <-listener.done:
		return nil, errors.New("Listener is closed")
	}
}

func (listener *Listener) Close() error {
	close(listener.done)
	return listener.inner.Close()
}

func (listener *Listener) Addr() net.Addr {
	return listener.inner.Addr()
}

func (listener *Listener) handleAccepts() {
	for {
		conn, err := listener.inner.Accept()
		if err != nil {
			if listener.sendError(err) {
				continue
			} else {
				break
			}
		}
		go listener.handleConnection(conn)
	}
}

func (listener *Listener) handleConnection(conn net.Conn) {
	if err := conn.SetReadDeadline(time.Now().Add(1 * time.Minute)); err != nil {
		conn.Close()
		listener.sendError(&acceptError{error: err, temporary: true})
		return
	}

	header, err := ReadHeader(conn)
	if err != nil {
		conn.Close()
		err = fmt.Errorf("Reading proxy header: %w", err)
		listener.sendError(&acceptError{error: err, temporary: true})
		return
	}

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		conn.Close()
		listener.sendError(&acceptError{error: err, temporary: true})
		return
	}

	proxyConn := &proxyConn{
		Conn:       conn,
		localAddr:  header.LocalAddr,
		remoteAddr: header.RemoteAddr,
	}
	if !listener.sendConn(proxyConn) {
		proxyConn.Close()
	}
}

func (listener *Listener) sendError(err error) bool {
	select {
	case listener.errors <- err:
		return true
	case <-listener.done:
		return false
	}
}

func (listener *Listener) sendConn(conn net.Conn) bool {
	select {
	case listener.conns <- conn:
		return true
	case <-listener.done:
		return false
	}
}
