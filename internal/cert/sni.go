package cert

import (
	"crypto/tls"
)

func GetCertificateDefaultServerName(defaultServerName string, getCertificate GetCertificateFunc) GetCertificateFunc {
	return func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		if hello.ServerName == "" {
			newHello := *hello
			newHello.ServerName = defaultServerName
			return getCertificate(&newHello)
		}
		return getCertificate(hello)
	}
}
