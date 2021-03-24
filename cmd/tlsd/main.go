package main

import (
	"flag"
	"log"
	"net"
	"src.agwa.name/go-listener"
	"src.agwa.name/go-listener/cert"
	"src.agwa.name/go-listener/socketdir"
)

func main() {
	var (
		socketDirectory   = flag.String("sockets", "/var/tls", "Directory for backend sockets")
		defaultHostname   = flag.String("default-hostname", "", "Default hostname if client does not provide SNI")
		defaultProtocol   = flag.String("default-protocol", "", "Default protocol if client does not provide ALPN")
		certDirectoryFlag = flag.String("certs", "/var/lib/certs", "Directory containing certificate bundles with the name SERVERNAME.pem")
		autocertFlag      = flag.Bool("autocert", false, "Obtain certificates automatically")
	)
	flag.Parse()

	ourListeners, err := listener.OpenAll(flag.Args())
	if err != nil {
		log.Fatal(err)
	}
	defer listener.CloseAll(ourListeners)

	server := &Server{
		SocketDirectory: &socketdir.Directory{Path: *socketDirectory},
		DefaultHostname: *defaultHostname,
		DefaultProtocol: *defaultProtocol,
	}

	if *autocertFlag {
		server.GetCertificate = cert.GetCertificateAutomatically(nil)
		server.HandleACME = true
	} else if *certDirectoryFlag != "" {
		server.GetCertificate = cert.GetCertificateFromDirectory(*certDirectoryFlag)
	}

	for _, listener := range ourListeners {
		go serve(listener, server)
	}

	select {}
}

func serve(listener net.Listener, server *Server) {
	log.Fatal(server.Serve(listener))
}
