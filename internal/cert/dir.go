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

func GetCertificateFromDirectory(path string) GetCertificateFunc {
	dir := &Directory{
		Path:  path,
		Cache: GlobalFileCache(),
	}
	return dir.GetCertificate
}
