package listener // import "src.agwa.name/go-listener"

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"src.agwa.name/go-listener/proxy"
)

func init() {
	RegisterListenerType("fd", openFDListener)
	RegisterListenerType("tcp", openTCPListener)
	RegisterListenerType("unix", openUnixListener)
	RegisterListenerType("proxy", openProxyListener)
}

func openFDListener(params map[string]interface{}, arg string) (net.Listener, error) {
	var fdString string
	if arg != "" {
		fdString = arg
	} else if param, ok := params["fd"].(json.Number); ok {
		fdString = string(param)
	} else if param, ok := params["fd"].(string); ok {
		fdString = param
	} else {
		return nil, errors.New("file descriptor not specified for FD listener")
	}

	fd, err := strconv.ParseUint(fdString, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("'%s' is a malformed file descriptor", fdString)
	}

	file := os.NewFile(uintptr(fd), fdString)
	defer file.Close()

	return net.FileListener(file)
}

func openTCPListener(params map[string]interface{}, arg string) (net.Listener, error) {
	var ipString string
	var portString string
	var err error

	if arg != "" {
		if strings.Contains(arg, ":") {
			ipString, portString, err = net.SplitHostPort(arg)
			if err != nil {
				return nil, fmt.Errorf("TCP listener has invalid argument: %w", err)
			}
		} else {
			portString = arg
		}
	} else if param, ok := params["address"].(string); ok {
		ipString = param
	} else if param, ok := params["port"].(json.Number); ok {
		portString = string(param)
	} else if param, ok := params["port"].(string); ok {
		portString = param
	}

	address := new(net.TCPAddr)

	if ipString != "" {
		address.IP = net.ParseIP(ipString)
		if address.IP == nil {
			return nil, errors.New("TCP listener has invalid IP address")
		}
	}

	address.Port, err = strconv.Atoi(portString)
	if err != nil {
		return nil, fmt.Errorf("TCP listener has invalid port: %w", err)
	}

	return net.ListenTCP("tcp", address)
}

func openUnixListener(params map[string]interface{}, arg string) (net.Listener, error) {
	var path string
	if arg != "" {
		path = arg
	} else if value, ok := params["path"].(string); ok {
		path = value
	} else {
		return nil, errors.New("path not specified for UNIX listener")
	}
	return net.ListenUnix("unix", &net.UnixAddr{Net: "unix", Name: path})
}

func openProxyListener(params map[string]interface{}, arg string) (net.Listener, error) {
	var inner net.Listener
	var err error
	if arg != "" {
		inner, err = Open(arg)
	} else if spec, ok := params["listener"].(map[string]interface{}); ok {
		inner, err = OpenJSON(spec)
	} else {
		return nil, errors.New("inner socket not specified for proxy listener")
	}
	if err != nil {
		return nil, err
	}
	return proxy.NewListener(inner), nil
}
