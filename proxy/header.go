package proxy

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

var protocolSignature = [12]byte{0x0D, 0x0A, 0x0D, 0x0A, 0x00, 0x0D, 0x0A, 0x51, 0x55, 0x49, 0x54, 0x0A}

const (
	commandLocal = 0x00
	commandProxy = 0x01
)

const (
	familyUnspecified = 0x00
	familyTCP4        = 0x11
	familyUDP4        = 0x12
	familyTCP6        = 0x21
	familyUDP6        = 0x22
)

type Header struct {
	RemoteAddr net.Addr
	LocalAddr  net.Addr
}

func ReadHeader(conn net.Conn) (*Header, error) {
	var preamble [16]byte
	if _, err := io.ReadFull(conn, preamble[:]); err != nil {
		return nil, err
	}

	var (
		signature = preamble[0:12]
		command   = preamble[12]
		family    = preamble[13]
		length    = binary.BigEndian.Uint16(preamble[14:16])
	)

	if !bytes.Equal(signature[:], protocolSignature[:]) {
		return nil, errors.New("Not a proxied connection")
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return nil, err
	}

	switch command {
	case commandLocal:
		return &Header{LocalAddr: conn.LocalAddr(), RemoteAddr: conn.RemoteAddr()}, nil
	case commandProxy:
		return parseProxyHeader(family, payload)
	default:
		return nil, fmt.Errorf("Unsupported proxy command %x", command)
	}
}

func parseProxyHeader(family uint8, payload []byte) (*Header, error) {
	switch family {
	case familyTCP4:
		if len(payload) < 12 {
			return nil, errors.New("Header too short for TCP over IPv4")
		}
		return &Header{
			RemoteAddr: &net.TCPAddr{
				IP:   payload[0:4],
				Port: int(binary.BigEndian.Uint16(payload[8:10])),
			},
			LocalAddr: &net.TCPAddr{
				IP:   payload[4:8],
				Port: int(binary.BigEndian.Uint16(payload[10:12])),
			},
		}, nil
	case familyUDP4:
		if len(payload) < 12 {
			return nil, errors.New("Header too short for UDP over IPv4")
		}
		return &Header{
			RemoteAddr: &net.UDPAddr{
				IP:   payload[0:4],
				Port: int(binary.BigEndian.Uint16(payload[8:10])),
			},
			LocalAddr: &net.UDPAddr{
				IP:   payload[4:8],
				Port: int(binary.BigEndian.Uint16(payload[10:12])),
			},
		}, nil
	case familyTCP6:
		if len(payload) < 36 {
			return nil, errors.New("Header too short for TCP over IPv6")
		}
		return &Header{
			RemoteAddr: &net.TCPAddr{
				IP:   payload[0:16],
				Port: int(binary.BigEndian.Uint16(payload[32:34])),
			},
			LocalAddr: &net.TCPAddr{
				IP:   payload[16:32],
				Port: int(binary.BigEndian.Uint16(payload[34:36])),
			},
		}, nil
	case familyUDP6:
		if len(payload) < 36 {
			return nil, errors.New("Header too short for UDP over IPv6")
		}
		return &Header{
			RemoteAddr: &net.UDPAddr{
				IP:   payload[0:16],
				Port: int(binary.BigEndian.Uint16(payload[32:34])),
			},
			LocalAddr: &net.UDPAddr{
				IP:   payload[16:32],
				Port: int(binary.BigEndian.Uint16(payload[34:36])),
			},
		}, nil
	default:
		return nil, fmt.Errorf("Unsupported address family %x", family)
	}
}

func (header Header) Format() []byte {
	switch remoteAddr := header.RemoteAddr.(type) {
	case *net.TCPAddr:
		localAddr := header.LocalAddr.(*net.TCPAddr)
		if remoteAddr.IP.To4() != nil {
			return formatIPv4Header(familyTCP4, remoteAddr.IP, localAddr.IP, remoteAddr.Port, localAddr.Port)
		} else {
			return formatIPv6Header(familyTCP6, remoteAddr.IP, localAddr.IP, remoteAddr.Port, localAddr.Port)
		}
	case *net.UDPAddr:
		localAddr := header.LocalAddr.(*net.UDPAddr)
		if remoteAddr.IP.To4() != nil {
			return formatIPv4Header(familyUDP4, remoteAddr.IP, localAddr.IP, remoteAddr.Port, localAddr.Port)
		} else {
			return formatIPv6Header(familyUDP6, remoteAddr.IP, localAddr.IP, remoteAddr.Port, localAddr.Port)
		}
	default:
		return formatUnspecifiedHeader()
	}
}

func formatIPv4Header(family uint8, remoteIP, localIP net.IP, remotePort, localPort int) []byte {
	header := make([]byte, 28)
	copy(header[0:12], protocolSignature[:])
	header[12] = commandProxy
	header[13] = family
	binary.BigEndian.PutUint16(header[14:16], 12)
	copy(header[16:20], remoteIP.To4())
	copy(header[20:24], localIP.To4())
	binary.BigEndian.PutUint16(header[24:26], uint16(remotePort))
	binary.BigEndian.PutUint16(header[26:28], uint16(localPort))
	return header[:]
}

func formatIPv6Header(family uint8, remoteIP, localIP net.IP, remotePort, localPort int) []byte {
	header := make([]byte, 52)
	copy(header[0:12], protocolSignature[:])
	header[12] = commandProxy
	header[13] = family
	binary.BigEndian.PutUint16(header[14:16], 36)
	copy(header[16:32], remoteIP)
	copy(header[32:48], localIP)
	binary.BigEndian.PutUint16(header[48:50], uint16(remotePort))
	binary.BigEndian.PutUint16(header[50:52], uint16(localPort))
	return header[:]
}

func formatUnspecifiedHeader() []byte {
	var header [16]byte
	copy(header[0:12], protocolSignature[:])
	header[12] = commandProxy
	header[13] = familyUnspecified
	return header[:]
}
