package cert

import (
	"crypto/tls"
)

func GetCertificateForFile(path string) func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	cache := GlobalFileCache()

	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		return cache.Load(path)
	}
}
