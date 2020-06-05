package listeners // import "src.agwa.name/go-listeners"

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

func Open(listenerType string, params map[string]interface{}, argument string) (net.Listener, error) {
	openListener := getOpenListenerFunc(listenerType)
	if openListener == nil {
		return nil, fmt.Errorf("Unknown listener type: " + listenerType)
	}
	return openListener(params, argument)
}

func OpenFromSpec(spec string) (net.Listener, error) {
	if strings.Contains(spec, ":") {
		fields := strings.SplitN(spec, ":", 2)
		listenerType, arg := fields[0], fields[1]
		return Open(listenerType, nil, arg)
	} else {
		return openTCPListener(nil, spec)
	}
}

func OpenFromSpecs(specs string) ([]net.Listener, error) {
	listeners := []net.Listener{}
	for _, spec := range strings.Split(specs, ",") {
		listener, err := OpenFromSpec(spec)
		if err != nil {
			CloseAll(listeners)
			return nil, fmt.Errorf("%s: %w", spec, err)
		}
		listeners = append(listeners, listener)
	}
	return listeners, nil
}

func OpenFromJSON(spec map[string]interface{}) (net.Listener, error) {
	listenerType, ok := spec["type"].(string)
	if !ok {
		return nil, errors.New("Listener object does not contain a string type field")
	}
	return Open(listenerType, spec, "")
}
