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
