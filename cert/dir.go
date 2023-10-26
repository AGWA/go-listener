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

type directory struct {
	Path  string
	Cache *fileCache
}

func (dir *directory) loadCertificate(filename string) (*tls.Certificate, error) {
	fullpath := filepath.Join(dir.Path, filename)

	if dir.Cache != nil {
		return dir.Cache.Load(fullpath)
	} else {
		return LoadCertificate(fullpath)
	}
}

func (dir *directory) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	serverName := hello.ServerName
	if serverName == "" {
		return nil, errors.New("Client does not support SNI")
	}
	if serverName[0] == '.' || strings.IndexByte(hello.ServerName, '/') != -1 {
		return nil, errors.New("Server name is invalid")
	}
	var serverNameSuffix string
	if dot := strings.IndexByte(serverName, '.'); dot != -1 {
		serverNameSuffix = serverName[dot:]
	}
	keyTypes := getSupportedKeyTypes(hello)

	if keyTypes.ecdsa {
		if cert, err := dir.loadCertificate(serverName + ".pem.ecdsa"); err == nil {
			return cert, nil
		} else if !os.IsNotExist(err) {
			// TODO: log this
		}
		if cert, err := dir.loadCertificate("_" + serverNameSuffix + ".pem.ecdsa"); err == nil {
			return cert, nil
		} else if !os.IsNotExist(err) {
			// TODO: log this
		}
	}

	if keyTypes.rsa {
		if cert, err := dir.loadCertificate(serverName + ".pem.rsa"); err == nil {
			return cert, nil
		} else if !os.IsNotExist(err) {
			// TODO: log this
		}
		if cert, err := dir.loadCertificate("_" + serverNameSuffix + ".pem.rsa"); err == nil {
			return cert, nil
		} else if !os.IsNotExist(err) {
			// TODO: log this
		}
	}

	if cert, err := dir.loadCertificate(serverName + ".pem"); err == nil {
		return cert, nil
	} else if !os.IsNotExist(err) {
		// TODO: log this
	}

	if cert, err := dir.loadCertificate("_" + serverNameSuffix + ".pem"); err == nil {
		return cert, nil
	} else if !os.IsNotExist(err) {
		// TODO: log this
	}

	return nil, errors.New("No certificate found")
}

// Return a [GetCertificateFunc] that gets the certificate from a file
// in the given directory.  The function searches for files in the
// following order:
//
//  1. SERVER_NAME.pem.ecdsa (only if client supports ECDSA certificates)
//  2. WILDCARD_NAME.pem.ecdsa (only if client supports ECDSA certificates)
//  3. SERVER_NAME.pem.rsa (only if client supports RSA certificates)
//  4. WILDCARD_NAME.pem.rsa (only if client supports RSA certificates)
//  5. SERVER_NAME.pem
//  6. WILDCARD_NAME.pem
//
// SERVER_NAME is the SNI hostname provided by the client, and WILDCARD_NAME
// is the SNI hostname with the first label replaced with an underscore
// (e.g. the wildcard name for www.example.com is _.example.com)
//
// Certificate files are cached in memory, and reloaded automatically when they
// change, allowing zero-downtime certificate rotation.  See the documentation of
// [LoadCertificate] for the required format of the files.
//
// If no certificate file is found, or if the client does not
// provide an SNI hostname, then the GetCertificateFunc returns an error,
// causing the TLS connection to be terminated.  If you need to support clients
// that don't provide SNI, wrap the GetCertificateFunc with
// [GetCertificateDefaultServerName] to specify a default SNI hostname.
func GetCertificateFromDirectory(path string) GetCertificateFunc {
	dir := &directory{
		Path:  path,
		Cache: globalFileCache(),
	}
	return dir.GetCertificate
}
