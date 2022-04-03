// Copyright (C) 2022 Andrew Ayer
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.
//
// Except as contained in this notice, the name(s) of the above copyright
// holders shall not be used in advertising or otherwise to promote the
// sale, use or other dealings in this Software without prior written
// authorization.

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
