package cert

import (
	"crypto/tls"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type CertificateDirectory struct {
	DefaultServerName string

	path  string
	mu    sync.RWMutex
	cache map[string]cachedCertificate
}

func NewCertificateDirectory(path string) *CertificateDirectory {
	return &CertificateDirectory{
		path:  path,
		cache: make(map[string]cachedCertificate),
	}
}

func (dir *CertificateDirectory) getCached(filename string) cachedCertificate {
	dir.mu.RLock()
	defer dir.mu.RUnlock()
	return dir.cache[filename]
}

func (dir *CertificateDirectory) addToCache(filename string, cert cachedCertificate) {
	dir.mu.Lock()
	defer dir.mu.Unlock()
	dir.cache[filename] = cert
}

func (dir *CertificateDirectory) LoadCertificate(filename string) (*tls.Certificate, error) {
	fileinfo, err := os.Stat(filepath.Join(dir.path, filename))
	if err != nil {
		return nil, err
	}
	modTime := fileinfo.ModTime()

	if cachedCert := dir.getCached(filename); cachedCert.isFresh(modTime) {
		return cachedCert.Certificate, nil
	}
	cert, err := LoadCertificate(filename)
	if err != nil {
		return nil, err
	}
	dir.addToCache(filename, cachedCertificate{Certificate: cert, modTime: modTime})
	return cert, nil
}

func (dir *CertificateDirectory) CleanCache() {
	dir.mu.Lock()
	defer dir.mu.Unlock()

	now := time.Now()
	for filename, cert := range dir.cache {
		if now.After(cert.Leaf.NotAfter) {
			delete(dir.cache, filename)
		}
	}
}

func (dir *CertificateDirectory) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	serverName := hello.ServerName
	if serverName == "" {
		if dir.DefaultServerName == "" {
			return nil, errors.New("Client did not provide SNI and DefaultServerName is not set")
		}
		serverName = dir.DefaultServerName
	}

	if serverName[0] == '.' || strings.IndexByte(hello.ServerName, '/') != -1 {
		return nil, errors.New("Server name is invalid")
	}

	if cert, err := dir.LoadCertificate(serverName + ".pem"); err == nil {
		return cert, nil
	} else if !os.IsNotExist(err) {
		// TODO: log this
	}

	serverName = replaceFirstLabel(serverName, "_")
	if cert, err := dir.LoadCertificate(serverName + ".pem"); err == nil {
		return cert, nil
	} else if !os.IsNotExist(err) {
		// TODO: log this
	}

	return nil, errors.New("No certificate found")
}
