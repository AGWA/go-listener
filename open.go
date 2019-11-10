package listeners // import "src.agwa.name/go-listeners"

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func openTCPListenerForAddress(addressString string) (*net.TCPListener, error) {
	address, err := net.ResolveTCPAddr("tcp", addressString)
	if err != nil {
		return nil, err
	}
	return net.ListenTCP("tcp", address)
}

func openUDPListenerForAddress(addressString string) (*net.UDPConn, error) {
	address, err := net.ResolveUDPAddr("udp", addressString)
	if err != nil {
		return nil, err
	}
	return net.ListenUDP("udp", address)
}

func openTCPListenerForFileDesc(fdString string) (*net.TCPListener, error) {
	fd, err := strconv.ParseUint(fdString, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("'%s' is a malformed file descriptor", fdString)
	}
	file := os.NewFile(uintptr(fd), fdString)
	fileListener, err := net.FileListener(file)
	file.Close()
	if err != nil {
		return nil, err
	}
	listener, isTCP := fileListener.(*net.TCPListener)
	if !isTCP {
		fileListener.Close()
		return nil, fmt.Errorf("%s is not a TCP listener", fdString)
	}
	return listener, nil
}

func openUDPListenerForFileDesc(fdString string) (*net.UDPConn, error) {
	fd, err := strconv.ParseUint(fdString, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("'%s' is a malformed file descriptor", fdString)
	}
	file := os.NewFile(uintptr(fd), fdString)
	fileListener, err := net.FilePacketConn(file)
	file.Close()
	if err != nil {
		return nil, err
	}
	listener, isUDP := fileListener.(*net.UDPConn)
	if !isUDP {
		fileListener.Close()
		return nil, fmt.Errorf("%s is not a UDP listener", fdString)
	}
	return listener, nil
}

func OpenTCP(addresses string, filedescs string) ([]*net.TCPListener, error) {
	listeners := []*net.TCPListener{}
	if addresses != "" {
		for _, addressString := range strings.Split(addresses, ",") {
			listener, err := openTCPListenerForAddress(addressString)
			if err != nil {
				CloseTCP(listeners)
				return nil, err
			}
			listeners = append(listeners, listener)
		}
	}
	if filedescs != "" {
		for _, fdString := range strings.Split(filedescs, ",") {
			listener, err := openTCPListenerForFileDesc(fdString)
			if err != nil {
				CloseTCP(listeners)
				return nil, err
			}
			listeners = append(listeners, listener)
		}
	}
	return listeners, nil
}

func OpenUDP(addresses string, filedescs string) ([]*net.UDPConn, error) {
	listeners := []*net.UDPConn{}
	if addresses != "" {
		for _, addressString := range strings.Split(addresses, ",") {
			listener, err := openUDPListenerForAddress(addressString)
			if err != nil {
				CloseUDP(listeners)
				return nil, err
			}
			listeners = append(listeners, listener)
		}
	}
	if filedescs != "" {
		for _, fdString := range strings.Split(filedescs, ",") {
			listener, err := openUDPListenerForFileDesc(fdString)
			if err != nil {
				CloseUDP(listeners)
				return nil, err
			}
			listeners = append(listeners, listener)
		}
	}
	return listeners, nil
}
