package socketdir

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
)

type Directory struct {
	Path string
}

func (dir *Directory) getHostDir(hostname string) string {
	hostname = canonicalizeHostname(hostname)
	if dir := filepath.Join(dir.Path, hostname); directoryExists(dir) {
		return dir
	}
	if dir := filepath.Join(dir.Path, replaceFirstLabel(hostname, "_")); directoryExists(dir) {
		return dir
	}
	return ""
}

func (dir *Directory) getBackendPath(hostname string, service string) (string, ProxyProto) {
	hostDir := dir.getHostDir(hostname)
	if hostDir == "" {
		return "", ""
	}

	backendDir := filepath.Join(hostDir, url.PathEscape(service))

	for _, proxyProto := range proxyProtos {
		sockPath := filepath.Join(backendDir, string(proxyProto))
		info, err := os.Stat(sockPath)
		if err == nil && (info.Mode()&os.ModeSocket) != 0 {
			return sockPath, proxyProto
		} else if err == nil {
			//log.Printf("Ignoring %s because it is not a socket file", sockPath)
		} else if !os.IsNotExist(err) {
			//log.Printf("Ignoring %s due to stat error: %s", sockPath, err)
		}
	}

	return "", ""
}

func (dir *Directory) ServesHostname(hostname string) bool {
	hostDir := dir.getHostDir(hostname)
	return hostDir != ""
}

func (dir *Directory) HasBackend(hostname string, service string) bool {
	path, _ := dir.getBackendPath(hostname, service)
	return path != ""
}

func (dir *Directory) DialBackend(hostname string, service string) (Conn, ProxyProto, error) {
	backendPath, proxyProto := dir.getBackendPath(hostname, service)
	if backendPath == "" {
		return nil, "", fmt.Errorf("cannot dial non-existent backend for host %q, service %q", hostname, service)
	}

	backendConn, err := net.DialUnix("unix", nil, &net.UnixAddr{Net: "unix", Name: backendPath})
	if err != nil {
		return nil, "", fmt.Errorf("error dialing backend for host %q, service %q: %w", hostname, service, err)
	}
	return backendConn, proxyProto, nil
}

func (dir *Directory) ProxyToBackend(hostname string, service string, clientConn Conn) error {
	backendConn, proxyProto, err := dir.DialBackend(hostname, service)
	if err != nil {
		return err
	}
	defer backendConn.Close()

	ProxyConnection(proxyProto, clientConn, backendConn)
	return nil
}
