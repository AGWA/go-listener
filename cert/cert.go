package cert

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

type GetCertificateFunc func(*tls.ClientHelloInfo) (*tls.Certificate, error)

func LoadCertificate(filename string) (*tls.Certificate, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cert := new(tls.Certificate)
	for {
		block, rest := pem.Decode(data)
		if block == nil {
			break
		}
		switch block.Type {
		case "PRIVATE KEY", "RSA PRIVATE KEY", "EC PRIVATE KEY":
			if cert.PrivateKey != nil {
				return nil, errors.New("contains more than one private key")
			}
			cert.PrivateKey, err = parsePrivateKey(block)
			if err != nil {
				return nil, fmt.Errorf("contains invalid private key: %w")
			}
		case "CERTIFICATE":
			cert.Certificate = append(cert.Certificate, block.Bytes)
		case "OCSP RESPONSE":
			if cert.OCSPStaple != nil {
				return nil, errors.New("contains more than one OCSP response")
			}
			cert.OCSPStaple = block.Bytes
		case "SIGNED CERTIFICATE TIMESTAMP":
			cert.SignedCertificateTimestamps = append(cert.SignedCertificateTimestamps, block.Bytes)
		default:
			return nil, errors.New("contains unrecognized PEM block `" + block.Type + "'")
		}
		data = rest
	}

	if cert.PrivateKey == nil {
		return nil, errors.New("doesn't contain any private key")
	}
	if len(cert.Certificate) == 0 {
		return nil, errors.New("doesn't contain any certificates")
	}
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("contains invalid leaf certificate: %w")
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
