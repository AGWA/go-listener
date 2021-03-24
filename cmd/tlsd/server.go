package main

import (
	"crypto/tls"
	"golang.org/x/crypto/acme"
	"log"
	"io"
	"net"
	"errors"
	"fmt"
	"time"

	"src.agwa.name/go-listener/socketdir"
	"src.agwa.name/go-listener/proxy"
	"src.agwa.name/go-listener/tlsutil"
)

type Server struct {
	GetCertificate  func(*tls.ClientHelloInfo) (*tls.Certificate, error)
	HandleACME      bool
	SocketDirectory *socketdir.Directory
	DefaultHostname string
	DefaultProtocol string
}

func (server *Server) handleACMEConnection(clientConn net.Conn, clientHello *tls.ClientHelloInfo) {
	if !server.SocketDirectory.ServesHostname(clientHello.ServerName) {
		log.Print("Ignoring ACME connection from %s because we don't serve %q", clientConn.RemoteAddr(), clientHello.ServerName)
		return
	}

	if err := clientConn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		log.Print("Error during ACME connection from %s: %s", clientConn.RemoteAddr(), err)
		return
	}

	err := tls.Server(clientConn, &tls.Config{
		GetCertificate: server.GetCertificate,
		NextProtos:     []string{acme.ALPNProto},
	}).Handshake()

	if err != nil {
		log.Print("TLS handshake for ACME connection from %s failed: %s", clientConn.RemoteAddr(), err)
		return
	}
}

func (server *Server) peekClientHello(clientConn net.Conn) (*tls.ClientHelloInfo, net.Conn, error) {
	if err := clientConn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return nil, nil, err
	}

	clientHello, peekedClientConn, err := tlsutil.PeekClientHelloFromConn(clientConn)
	if err != nil {
		return nil, nil, err
	}

	if err := clientConn.SetReadDeadline(time.Time{}); err != nil {
		return nil, nil, err
	}

	if clientHello.ServerName == "" {
		if server.DefaultHostname == "" {
			return nil, nil, errors.New("no SNI provided and DefaultHostname not set")
		}
		clientHello.ServerName = server.DefaultHostname
	}
	if len(clientHello.SupportedProtos) == 0 {
		if server.DefaultProtocol == "" {
			return nil, nil, errors.New("no ALPN provided and DefaultProtocol not set")
		}
		clientHello.SupportedProtos = []string{server.DefaultProtocol}
	}

	return clientHello, peekedClientConn, err
}

func (server *Server) terminateTLS(clientConn net.Conn, clientHello *tls.ClientHelloInfo, protocol string) (*tls.Conn, error) {
	if server.GetCertificate == nil {
		return nil, errors.New("certificate source not configured")
	}

	cert, err := server.GetCertificate(clientHello)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	if err := clientConn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return nil, err
	}

	tlsConn := tls.Server(clientConn, &tls.Config{
		Certificates: []tls.Certificate{*cert},
		NextProtos:   []string{protocol},
	})

	if err := tlsConn.Handshake(); err != nil {
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}

	if err := clientConn.SetDeadline(time.Time{}); err != nil {
		return nil, err
	}

	return tlsConn, nil
}

func (server *Server) handleConnection(clientConn net.Conn) {
	defer func() { clientConn.Close() }()

	var clientHello *tls.ClientHelloInfo

	if peekedClientHello, peekedClientConn, err := server.peekClientHello(clientConn); err == nil {
		clientHello = peekedClientHello
		clientConn = peekedClientConn
	} else {
		log.Printf("Peeking client hello from %s failed: %s", clientConn.RemoteAddr(), err)
		return
	}

	if server.HandleACME && len(clientHello.SupportedProtos) == 1 && clientHello.SupportedProtos[0] == acme.ALPNProto {
		server.handleACMEConnection(clientConn, clientHello)
		return
	}

	backend := server.SocketDirectory.GetBackend(clientHello.ServerName, clientHello.SupportedProtos)
	if backend == nil {
		log.Printf("Ignoring connection from %s because we don't serve %q for %q", clientConn.RemoteAddr(), clientHello.ServerName, clientHello.SupportedProtos)
		return
	}

	if !backend.Type.TLS {
		// since the backend is _not_ a TLS backend, we need to terminate TLS for it
		tlsConn, err := server.terminateTLS(clientConn, clientHello, backend.Service)
		if err != nil {
			log.Printf("Terminating TLS connection from %s failed: %s", clientConn.RemoteAddr(), err)
			return
		}
		clientConn = tlsConn
	}

	backendConn, err := server.SocketDirectory.Dial(backend)
	if err != nil {
		log.Printf("Ignoring connection from %s because dialing backend failed: %s", clientConn.RemoteAddr(), err)
		return
	}
	defer backendConn.Close()

	if backend.Type.ProxyProto {
		// TODO: consider also sending hostname and protocol using PP2_TYPE_AUTHORITY and PP2_TYPE_ALPN... we could also send detailed TLS info using PP2_TYPE_SSL
		header := proxy.Header{RemoteAddr: clientConn.RemoteAddr(), LocalAddr: clientConn.LocalAddr()}
		if _, err := backendConn.Write(header.Format()); err != nil {
			log.Printf("Error writing PROXY header to backend: %s", err)
			return
		}
	}

	go func() {
		io.Copy(backendConn, clientConn)
		backendConn.CloseWrite()
	}()

	io.Copy(clientConn, backendConn)
}

func (server *Server) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if netErr, isNetErr := err.(net.Error); isNetErr && netErr.Temporary() {
				log.Printf("Temporary network error accepting connection: %s", netErr)
				continue
			}
			return err
		}
		go server.handleConnection(conn)
	}
}
