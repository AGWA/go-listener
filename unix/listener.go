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

// Package unix implements a net.Listener for UNIX domain sockets.
package unix // import "src.agwa.name/go-listener/unix"

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

type watchedListener struct {
	listener *net.UnixListener
	closed   chan struct{}
}

func newWatchedListener(listener *net.UnixListener) *watchedListener {
	return &watchedListener{
		listener: listener,
		closed:   make(chan struct{}),
	}
}

func (wl *watchedListener) Accept() (net.Conn, error) {
	return wl.listener.Accept()
}

func (wl *watchedListener) Close() error {
	close(wl.closed)
	return wl.listener.Close()
}

func (wl *watchedListener) Addr() net.Addr {
	return wl.listener.Addr()
}

func (wl *watchedListener) watch(path string, info os.FileInfo) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-wl.closed:
			return
		case <-ticker.C:
			latestInfo, err := os.Lstat(path)
			if !(err == nil && os.SameFile(info, latestInfo)) {
				wl.listener.Close()
				return
			}
		}
	}
}

// Create a listening UNIX domain socket with the given path and filesystem
// permissions.  If a file already exists at the path, it is replaced.  If
// the UNIX domain socket file is removed or changed, then within 5 seconds
// the net.Listener will be closed, and Accept will return an error.
func Listen(path string, mode os.FileMode) (net.Listener, error) {
	tempDir, err := os.MkdirTemp(filepath.Dir(path), ".tmp")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary directory to hold Unix socket: %w")
	}
	defer os.Remove(tempDir)

	tempPath := filepath.Join(tempDir, "socket")
	tempListener, err := net.ListenUnix("unix", &net.UnixAddr{Net: "unix", Name: tempPath})
	if err != nil {
		return nil, err
	}
	tempListener.SetUnlinkOnClose(false)
	defer os.Remove(tempPath)
	defer func() {
		if tempListener != nil {
			tempListener.Close()
		}
	}()

	if err := os.Chmod(tempPath, mode); err != nil {
		return nil, err
	}

	fileInfo, err := os.Lstat(tempPath)
	if err != nil {
		return nil, err
	}

	if err := os.Rename(tempPath, path); err != nil {
		return nil, err
	}

	listener := newWatchedListener(tempListener)
	tempListener = nil

	go listener.watch(path, fileInfo)
	return listener, nil
}
