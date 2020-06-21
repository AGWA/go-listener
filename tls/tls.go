package tls

import (
	"crypto/tls"
	"errors"
	"net"
	"strings"

	"src.agwa.name/go-listener"
	"src.agwa.name/go-listener/cert"
)

func init() {
	listener.RegisterListenerType("tls", openTLSListener)
}

func openTLSListener(params map[string]interface{}, arg string) (net.Listener, error) {
	var getCertificate cert.GetCertificateFunc
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
	}

	return tls.NewListener(inner, config), nil
}
