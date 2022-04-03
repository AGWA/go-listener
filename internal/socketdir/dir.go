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

package socketdir

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Directory struct {
	Path string
}

func (dir *Directory) hostnamePath(hostname string) string {
	return filepath.Join(dir.Path, hostname)
}

func (dir *Directory) socketPath(hostname string, service string, backendType BackendType) string {
	return filepath.Join(dir.hostnamePath(hostname), url.PathEscape(service), backendType.socketFilename)
}

func (dir *Directory) hostnameDirExists(hostname string) bool {
	hostnamePath := dir.hostnamePath(hostname)
	info, err := os.Stat(hostnamePath)
	if err == nil && info.IsDir() {
		return true
	}
	if err == nil {
		//log.Printf("Ignoring %s because it is not a directory", hostnamePath)
	} else if !os.IsNotExist(err) {
		//log.Printf("Ignoring %s due to stat error: %s", hostnamePath, err)
	}
	return false
}

func (dir *Directory) canonicalizeHostname(hostname string) string {
	if len(hostname) == 0 || hostname[0] == '.' || strings.ContainsRune(hostname, '/') {
		return ""
	}

	hostname = strings.ToLower(hostname)
	hostname = strings.TrimRight(hostname, ".")

	if dir.hostnameDirExists(hostname) {
		return hostname
	}

	if wildcardHostname := replaceFirstLabel(hostname, "_"); dir.hostnameDirExists(wildcardHostname) {
		return wildcardHostname
	}

	return ""
}

func (dir *Directory) ServesHostname(hostname string) bool {
	return dir.canonicalizeHostname(hostname) != ""
}

func (dir *Directory) GetBackend(hostname string, services []string) *Backend {
	hostname = dir.canonicalizeHostname(hostname)
	if hostname == "" {
		return nil
	}

	for _, service := range services {
		for _, backendType := range backendTypes {
			socketPath := dir.socketPath(hostname, service, backendType)
			info, err := os.Stat(socketPath)
			if err == nil && (info.Mode()&os.ModeSocket) != 0 {
				return &Backend{Hostname: hostname, Service: service, Type: backendType}
			} else if err == nil {
				//log.Printf("Ignoring %s because it is not a socket file", socketPath)
			} else if !os.IsNotExist(err) {
				//log.Printf("Ignoring %s due to stat error: %s", socketPath, err)
			}
		}
	}
	return nil
}

func (dir *Directory) Dial(backend *Backend) (Conn, error) {
	socketPath := dir.socketPath(backend.Hostname, backend.Service, backend.Type)

	// TODO: consider setting a timeout on the dial
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Net: "unix", Name: socketPath})
	if err != nil {
		return nil, fmt.Errorf("dialing backend for host %q, service %q, type %q failed: %w", backend.Hostname, backend.Service, backend.Type.socketFilename, err)
	}

	return conn, nil
}
