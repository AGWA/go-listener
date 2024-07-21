// Copyright (C) 2022 Andrew Ayer
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.
//
// Except as contained in this notice, the name(s) of the above copyright
// holders shall not be used in advertising or otherwise to promote the
// sale, use or other dealings in this Software without prior written
// authorization.

// Package proxy implements version 2 of the PROXY protocol (https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt)
package proxy // import "src.agwa.name/go-listener/proxy"

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

type proxyListener struct {
	inner   net.Listener
	conns   chan net.Conn
	errors  chan error
	done    chan struct{}
	closeMu sync.Mutex
}

// NewListener creates a [net.Listener] which accepts connections from an
// inner net.Listener, reads the PROXY v2 header from the client, and
// sets the local and remote addresses of the [net.Conn] to the values
// specified in the PROXY header.
func NewListener(inner net.Listener) net.Listener {
	listener := &proxyListener{
		inner:  inner,
		conns:  make(chan net.Conn),
		errors: make(chan error),
		done:   make(chan struct{}),
	}
	go listener.handleAccepts()
	return listener
}

func (listener *proxyListener) Accept() (net.Conn, error) {
	select {
	case conn := <-listener.conns:
		return conn, nil
	case err := <-listener.errors:
		return nil, err
	case <-listener.done:
		return nil, net.ErrClosed
	}
}

func (listener *proxyListener) Close() error {
	listener.closeMu.Lock()
	defer listener.closeMu.Unlock()

	select {
	case <-listener.done:
		return net.ErrClosed
	default:
		close(listener.done)
		return listener.inner.Close()
	}
}

func (listener *proxyListener) Addr() net.Addr {
	return listener.inner.Addr()
}

func (listener *proxyListener) handleAccepts() {
	for {
		conn, err := listener.inner.Accept()
		if errors.Is(err, net.ErrClosed) {
			break
		} else if err != nil {
			if !listener.sendError(err) {
				break
			}
		} else {
			go listener.handleConnection(conn)
		}
	}
}

func (listener *proxyListener) handleConnection(conn net.Conn) {
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

func (listener *proxyListener) sendError(err error) bool {
	select {
	case listener.errors <- err:
		return true
	case <-listener.done:
		return false
	}
}

func (listener *proxyListener) sendConn(conn net.Conn) bool {
	select {
	case listener.conns <- conn:
		return true
	case <-listener.done:
		return false
	}
}
