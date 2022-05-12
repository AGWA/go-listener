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
