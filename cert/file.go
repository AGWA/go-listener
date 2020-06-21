package cert

import (
	"crypto/tls"
)

func GetCertificateFromFile(path string) GetCertificateFunc {
	cache := GlobalFileCache()

	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		return cache.Load(path)
	}
}
