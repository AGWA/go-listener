# `go-listener`

[![Documentation](https://pkg.go.dev/badge/src.agwa.name/go-listener)](https://pkg.go.dev/src.agwa.name/go-listener)

`src.agwa.name/go-listener` is a Go library for creating `net.Listener`s.

Typically, server software only supports listening on TCP ports.  `go-listener` makes it easy to also listen on:

* TCP ports
* UNIX domain sockets
* Pre-opened file descriptors

Additionally, `go-listener` makes it easy to support:

* The [PROXY protocol](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt)
* TLS (with several options for certificate management, including ACME)

Listeners are specified using a string syntax, which makes them easy to pass as command line arguments.

## How To Use

```go
import "src.agwa.name/go-listener"

netListener, err := listener.Open(listenerString)
if err != nil {
	// Handle err
}
defer netListener.Close()
```

`listener.Open` takes a string which describes a listener per the syntax described below, and returns a `net.Listener`, which you can use by calling `Accept`, passing to `http.Serve`, etc.

## Listener Syntax

### TCP

Listen on all interfaces:

```
tcp:PORT
```

Listen on a specific IPv4 interface:

```
tcp:IPV4ADDRESS:PORT
```

Listen on a specific IPv6 interface:

```
tcp:[IPV6ADDRESS]:PORT
```

Listen on all IPv4 interfaces:

```
tcp:0.0.0.0:PORT

```
Listen on all IPv6 interfaces:

```
tcp:[::]:PORT
```

### UNIX Domain Socket

```
unix:PATH
```

### File Descriptor

Listen on a file descriptor that is already open, bound, and listening:

```
fd:NUMBER
```

### PROXY Protocol

Wrap a listener with the [PROXY Protocol version 2](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt):

```
proxy:LISTENER
```

(where `LISTENER` is one of the syntaxes specified here)

`go-listener` will transparently read the PROXY protocol header and make the true client IP address available via the `net.Conn`'s `LocalAddr` method.

### TLS

Note: TLS support must be enabled by importing `src.agwa.name/go-listener/tls` like this:

```go
import _ "src.agwa.name/go-listener/tls"
```

Wrap a listener with TLS, using the certificate/key in the given file (which must be absolute path):

```
tls:/PATH/TO/CERTIFICATE_FILE:LISTENER
```

Wrap a listener with TLS, using the certificate/key named `SERVER_NAME.pem` in the given directory (which must be an absolute path and end with a slash):

```
tls:/PATH/TO/CERTIFICATE_DIRECTORY/:LISTENER
```

Wrap a listener with TLS and automatically obtain certificates for each hostname using ACME (requires the hostname to be publicly-accessible on port 443):

```
tls:HOSTNAME,HOSTNAME,...:LISTENER
```

#### Certificate Files

When you specify a certificate file or directory, certificates must be PEM-encoded and contain the following blocks:

* Exactly one `PRIVATE KEY`, containing the private key in PKCS#8 format.
* At least one `CERTIFICATE`, comprising the certificate chain, leaf certificate first and root certificate omitted.
* Up to one `OCSP RESPONSE`, containing a stapled OCSP response.
* Any number of `SIGNED CERTIFICATE TIMESTAMP`, containing stapled SCTs.

Certificate files are automatically reloaded when they change.

#### ACME Configuration

When you obtain certificates automatically, the following environment variables can be used to configure the ACME client:

| Environment Variable   | Description | Default |
| ---------------------- | ----------- | ------- |
| `AUTOCERT_ACME_SERVER` | The directory URL of the certificate authority's ACME server | [`autocert.DefaultACMEDirectory`](https://pkg.go.dev/golang.org/x/crypto/acme/autocert#DefaultACMEDirectory) |
| `AUTOCERT_EMAIL`       | Contact email address for your ACME account, used by certificate authority to notify you of certificate problems (highly recommended) | (none) |
| `AUTOCERT_EAB_KID`     | Key ID of the External Account Binding to use with ACME | (none) |
| `AUTOCERT_EAB_KEY`     | base64url-encoded HMAC-SHA256 key of the External Account Binding to use with ACME | (none) |
| `AUTOCERT_CACHE_DIR`   | The directory where issued certificates are stored | When root, `/var/lib/autocert-cache`; otherwise, `autocert-cache` under [`$XDG_DATA_HOME`](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) |

## Example

Here's how to use `go-listener` with `http.Server`:

```go
package main

import (
	"flag"
	"log"
	"net/http"

	"src.agwa.name/go-listener"
	_ "src.agwa.name/go-listener/tls"
)

func main() {
	var listenerString string
	flag.StringVar(&listenerString, "listen", "", "Socket to listen on")
	flag.Parse()

	netListener, err := listener.Open(listenerString)
	if err != nil {
		log.Fatal(err)
	}
	defer netListener.Close()

	log.Fatal(http.Serve(netListener, nil))
}

```

Listen on localhost, port 80:

```
httpd -listen tcp:127.0.0.1:80
```

Listen on IPv6 localhost, port 80:

```
httpd -listen tcp:[::1]:80
```

Listen on file descriptor 3:

```
httpd -listen fd:3
```

Listen on port 443, all interfaces, with TLS, using certificates in `/var/certs`:

```
httpd -listen tls:/var/certs/:tcp:443
```

Listen on port 443, all interfaces, with TLS, with automatic certificates for `www.example.com` and `example.com`:

```
httpd -listen tls:www.example.com,example.com:tcp:443
```

Listen on UNIX domain docket `/run/example.sock` with the PROXY protocol:

```
httpd -listen proxy:unix:/run/example.sock
```

Listen on UNIX domain socket `/run/example.sock` with TLS and the PROXY protocol, with certificate in `/etc/ssl/example.com.pem`:

```
httpd -listen tls:/etc/ssl/example.com.pem:proxy:unix:/run/example.sock
```

(Details: `go-listener` will listen on `/run/example.sock`.  When a connection is accepted, `go-listener` will first read the PROXY protocol header to get the true client IP address, which will be made available through the `net.Conn`'s `LocalAddr` method.  It will then do a TLS handshake using the private key and certificate in `/etc/ssl/example.com.pem`.)
