package cert

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"time"
)

type cachedCertificate struct {
	tls.Certificate
	modTime time.Time
}

func LoadCertificate(filename string) (tls.Certificate, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return tls.Certificate{}, err
	}

	var cert tls.Certificate
	for {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}
		switch block.Type {
		case "PRIVATE KEY", "RSA PRIVATE KEY", "EC PRIVATE KEY":
			if cert.PrivateKey != nil {
				return tls.Certificate{}, errors.New("contains more than one private key")
			}
			cert.PrivateKey, err = parsePrivateKey(block)
			if err != nil {
				return tls.Certificate{}, fmt.Errorf("contains invalid private key: %w")
			}
		case "CERTIFICATE":
			cert.Certificate = append(cert.Certificate, block.Bytes)
		case "OCSP RESPONSE":
			if cert.OCSPStaple != nil {
				return tls.Certificate{}, errors.New("contains more than one OCSP response")
			}
			cert.OCSPStaple = block.Bytes
		case "SIGNED CERTIFICATE TIMESTAMP":
			cert.SignedCertificateTimestamps = append(cert.SignedCertificateTimestamps, block.Bytes)
		default:
			return tls.Certificate{}, errors.New("contains unrecognized PEM block `" + block.Type + "'")
		}
		data = rest
	}

	if cert.PrivateKey == nil {
		return tls.Certificate{}, errors.New("doesn't contain any private key")
	}
	if len(cert.Certificate) == 0 {
		return tls.Certificate{}, errors.New("doesn't contain any certificates")
	}
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("contains invalid leaf certificate: %w")
	}

	return cert, nil
}

func parsePrivateKey(block *pem.Block) (crypto.PrivateKey, error) {
	switch block.Type {
	case "PRIVATE KEY":
		return x509.ParsePKCS8PrivateKey(block.Bytes)
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)
	default:
		return nil, errors.New("unrecognized private key type `" + block.Type + "'")
	}
}
