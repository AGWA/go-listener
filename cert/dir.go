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
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Directory struct {
	Path  string
	Cache *FileCache
}

func (dir *Directory) loadCertificate(filename string) (*tls.Certificate, error) {
	fullpath := filepath.Join(dir.Path, filename)

	if dir.Cache != nil {
		return dir.Cache.Load(fullpath)
	} else {
		return LoadCertificate(fullpath)
	}
}

func (dir *Directory) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	serverName := hello.ServerName
	if serverName == "" {
		return nil, errors.New("Client does not support SNI")
	}
	if serverName[0] == '.' || strings.IndexByte(hello.ServerName, '/') != -1 {
		return nil, errors.New("Server name is invalid")
	}

	if cert, err := dir.loadCertificate(serverName + ".pem"); err == nil {
		return cert, nil
	} else if !os.IsNotExist(err) {
		// TODO: log this
	}

	serverName = replaceFirstLabel(serverName, "_")
	if cert, err := dir.loadCertificate(serverName + ".pem"); err == nil {
		return cert, nil
	} else if !os.IsNotExist(err) {
		// TODO: log this
	}

	return nil, errors.New("No certificate found")
}

func replaceFirstLabel(hostname string, replacement string) string {
	dot := strings.IndexByte(hostname, '.')
	if dot == -1 {
		return replacement
	} else {
		return replacement + hostname[dot:]
	}
}

// Return a GetCertificateFunc that gets the certificate from a file
// named SERVER_NAME.pem in the given directory, where SERVER_NAME is
// the SNI hostname provided by the client.  File are reloaded automatically
// when they change, allowing zero-downtime certificate rotation.
// See the documentation of LoadCertificate for the required format of the files.
func GetCertificateFromDirectory(path string) GetCertificateFunc {
	dir := &Directory{
		Path:  path,
		Cache: GlobalFileCache(),
	}
	return dir.GetCertificate
}
