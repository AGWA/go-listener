package listener // import "src.agwa.name/go-listener"

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

func OpenType(listenerType string, params map[string]interface{}, argument string) (net.Listener, error) {
	openListener := getOpenListenerFunc(listenerType)
	if openListener == nil {
		return nil, fmt.Errorf("Unknown listener type: " + listenerType)
	}
	return openListener(params, argument)
}

func Open(spec string) (net.Listener, error) {
	if strings.Contains(spec, ":") {
		fields := strings.SplitN(spec, ":", 2)
		listenerType, arg := fields[0], fields[1]
		return OpenType(listenerType, nil, arg)
	} else {
		return openTCPListener(nil, spec)
	}
}

func OpenAll(specs []string) ([]net.Listener, error) {
	listeners := []net.Listener{}
	for _, spec := range specs {
		listener, err := Open(spec)
		if err != nil {
			CloseAll(listeners)
			return nil, fmt.Errorf("%s: %w", spec, err)
		}
		listeners = append(listeners, listener)
	}
	return listeners, nil
}

func OpenJSON(spec map[string]interface{}) (net.Listener, error) {
	listenerType, ok := spec["type"].(string)
	if !ok {
		return nil, errors.New("Listener object does not contain a string type field")
	}
	return OpenType(listenerType, spec, "")
}
