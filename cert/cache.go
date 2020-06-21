package cert

import (
	"crypto/tls"
	"os"
	"sync"
	"time"
)

type cachedFile struct {
	*tls.Certificate
	modTime time.Time
}

func (c *cachedFile) isFresh(latestModTime time.Time) bool {
	return c.Certificate != nil && c.modTime.Equal(latestModTime)
}

type FileCache struct {
	mu    sync.RWMutex
	certs map[string]cachedFile
}

func NewFileCache() *FileCache {
	return &FileCache{
		certs: make(map[string]cachedFile),
	}
}

func (c *FileCache) get(filename string) cachedFile {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.certs[filename]
}

func (c *FileCache) add(filename string, cert cachedFile) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.certs[filename] = cert
}

func (c *FileCache) Load(filename string) (*tls.Certificate, error) {
	fileinfo, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	modTime := fileinfo.ModTime()

	if cachedFile := c.get(filename); cachedFile.isFresh(modTime) {
		return cachedFile.Certificate, nil
	}
	cert, err := LoadCertificate(filename)
	if err != nil {
		return nil, err
	}
	c.add(filename, cachedFile{Certificate: cert, modTime: modTime})
	return cert, nil
}

func (c *FileCache) Clean() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for filename, cert := range c.certs {
		if now.After(cert.Leaf.NotAfter) {
			delete(c.certs, filename)
		}
	}
}
