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

// Package cert provides helper functions for working with TLS certificates.
package cert // import "src.agwa.name/go-listener/cert"

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

// A function that returns a [tls.Certificate] based on the given [tls.ClientHelloInfo]
type GetCertificateFunc func(*tls.ClientHelloInfo) (*tls.Certificate, error)

// Load a [tls.Certificate] from the given PEM-encoded file.  The file must contain
// the following blocks:
//   - Exactly one PRIVATE KEY, containing the private key in PKCS#8 format.
//   - At least one CERTIFICATE, comprising the certificate chain, leaf certificate first and root certificate omitted.
//   - Up to one OCSP RESPONSE, containing a stapled OCSP response.
//   - Any number of SIGNED CERTIFICATE TIMESTAMP, containing stapled SCTs.
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
			cert.PrivateKey, err = makePrivateKey(block)
			if err != nil {
				return nil, fmt.Errorf("contains invalid private key: %w", err)
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
		return nil, fmt.Errorf("contains invalid leaf certificate: %w", err)
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

func makePrivateKey(block *pem.Block) (crypto.PrivateKey, error) {
	key, err := parsePrivateKey(block)
	if err != nil {
		return nil, err
	}
	usage := block.Headers["Usage"]
	switch usage {
	case "":
		return key, nil
	case "decrypt":
		decrypter, ok := key.(crypto.Decrypter)
		if !ok {
			return nil, fmt.Errorf("this key type does not support decryption")
		}
		return struct{ crypto.Decrypter }{decrypter}, nil
	case "sign":
		signer, ok := key.(crypto.Signer)
		if !ok {
			return nil, fmt.Errorf("this key type does not support signing")
		}
		return struct{ crypto.Signer }{signer}, nil
	default:
		return nil, errors.New("unrecognized usage `" + usage + "'")
	}
}
