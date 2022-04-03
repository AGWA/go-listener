package listener // import "src.agwa.name/go-listener"

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

func openType(listenerType string, params map[string]interface{}, argument string) (net.Listener, error) {
	openListener := getOpenListenerFunc(listenerType)
	if openListener == nil {
		return nil, fmt.Errorf("Unknown listener type: " + listenerType)
	}
	return openListener(params, argument)
}

// Open a listener with the given string notation
func Open(spec string) (net.Listener, error) {
	if strings.Contains(spec, ":") {
		fields := strings.SplitN(spec, ":", 2)
		listenerType, arg := fields[0], fields[1]
		return openType(listenerType, nil, arg)
	} else {
		return openTCPListener(nil, spec)
	}
}

// Open all of the listeners specified in specs (using string notation).
// If any listener fails to open, an error is returned, and none of the
// listeners are opened.
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

// Open a listener with the given JSON notation.  Note that numbers in spec
// must be represented using json.Number.
func OpenJSON(spec map[string]interface{}) (net.Listener, error) {
	listenerType, ok := spec["type"].(string)
	if !ok {
		return nil, errors.New("Listener object does not contain a string type field")
	}
	return openType(listenerType, spec, "")
}
