// Copyright (C) 2023 Andrew Ayer
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

package listener // import "src.agwa.name/go-listener"

import (
	"errors"
	"net"
	"sync"
)

type multiAddr struct{}

func (multiAddr) Network() string { return "multi" }
func (multiAddr) String() string  { return "multi" }

type multiListener struct {
	listeners []net.Listener
	closed    chan struct{}
	conns     chan net.Conn
	errors    chan error
	closeMu   sync.Mutex
}

// Create a net.Listener that aggregates the provided listeners. Calling Accept() returns
// the next available connection among all the listeners.  Calling Close() closes each of
// the listeners, and causes blocked Accept calls to return with net.ErrClosed. Addr()
// returns a placeholder address that is probably not useful.
func MultiListener(listeners ...net.Listener) net.Listener {
	ml := &multiListener{
		listeners: listeners,
		closed:    make(chan struct{}),
		conns:     make(chan net.Conn),
		errors:    make(chan error),
	}
	for _, l := range ml.listeners {
		l := l
		go ml.handleAccepts(l)
	}
	return ml
}

func (ml *multiListener) handleAccepts(l net.Listener) {
	for {
		conn, err := l.Accept()
		if errors.Is(err, net.ErrClosed) {
			break
		} else if err != nil {
			if !ml.sendError(err) {
				break
			}
		} else {
			if !ml.sendConn(conn) {
				conn.Close()
				break
			}
		}
	}
}

func (ml *multiListener) sendError(err error) bool {
	select {
	case <-ml.closed:
		return false
	case ml.errors <- err:
		return true
	}
}

func (ml *multiListener) sendConn(conn net.Conn) bool {
	select {
	case <-ml.closed:
		return false
	case ml.conns <- conn:
		return true
	}
}

func (ml *multiListener) Accept() (net.Conn, error) {
	select {
	case <-ml.closed:
		return nil, net.ErrClosed
	case conn := <-ml.conns:
		return conn, nil
	case err := <-ml.errors:
		return nil, err
	}
}

func (ml *multiListener) Close() error {
	ml.closeMu.Lock()
	defer ml.closeMu.Unlock()

	select {
	case <-ml.closed:
		return net.ErrClosed
	default:
		close(ml.closed)
		for _, l := range ml.listeners {
			l.Close()
		}
		return nil
	}
}

func (ml *multiListener) Addr() net.Addr {
	return multiAddr{}
}
