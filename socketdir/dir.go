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
