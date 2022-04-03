// Package tls adds support for TLS listeners to src.agwa.name/go-listener.
//
// This package contains no exported identifiers and is intended to be
// imported from package main like this:
//
//     import _ "src.agwa.name/go-listener/tls"
//
package tls // import "src.agwa.name/go-listener/tls"

import (
	"crypto/tls"
	"errors"
	"net"
	"strings"

	"golang.org/x/crypto/acme"
	"src.agwa.name/go-listener"
	"src.agwa.name/go-listener/internal/cert"
)

func init() {
	listener.RegisterListenerType("tls", openHTTPSListener) // TODO: either remove this listener type or replace it with a generic TLS non-HTTPS listener
	listener.RegisterListenerType("https", openHTTPSListener)
}

func openHTTPSListener(params map[string]interface{}, arg string) (net.Listener, error) {
	var getCertificate cert.GetCertificateFunc
	var nextProtos = []string{"h2", "http/1.1"}
	var inner net.Listener
	var err error

	if arg != "" {
		fields := strings.SplitN(arg, ":", 2)
		if len(fields) < 2 {
			return nil, errors.New("TLS listener spec invalid; must be CERT_SPEC:SOCKET_SPEC")
		}
		certSpec, innerSpec := fields[0], fields[1]

		if strings.HasPrefix(certSpec, "/") && strings.HasSuffix(certSpec, "/") {
			getCertificate = cert.GetCertificateFromDirectory(certSpec)
		} else if strings.HasPrefix(certSpec, "/") {
			getCertificate = cert.GetCertificateFromFile(certSpec)
		} else {
			getCertificate = cert.GetCertificateAutomatically(strings.Split(certSpec, ","))
			nextProtos = append(nextProtos, acme.ALPNProto)
		}

		inner, err = listener.Open(innerSpec)
		if err != nil {
			return nil, err
		}
	} else {
		if path, ok := params["cert"].(string); ok {
			getCertificate = cert.GetCertificateFromFile(path)
		} else if path, ok := params["cert_directory"].(string); ok {
			getCertificate = cert.GetCertificateFromDirectory(path)
		} else if hostnames, ok := params["autocert_hostnames"].([]string); ok {
			getCertificate = cert.GetCertificateAutomatically(hostnames)
			nextProtos = append(nextProtos, acme.ALPNProto)
		} else {
			return nil, errors.New("certificate not specified for TLS listener")
		}

		innerSpec, ok := params["listener"].(map[string]interface{})
		if !ok {
			return nil, errors.New("inner socket not specified for TLS listener")
		}
		inner, err = listener.OpenJSON(innerSpec)
		if err != nil {
			return nil, err
		}
	}

	if defaultServerName, ok := params["default_server_name"].(string); ok && defaultServerName != "" {
		getCertificate = cert.GetCertificateDefaultServerName(defaultServerName, getCertificate)
	}

	config := &tls.Config{
		GetCertificate: getCertificate,
		NextProtos:     nextProtos,
	}

	return tls.NewListener(inner, config), nil
}
