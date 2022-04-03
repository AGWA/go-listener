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
	"net"
	"sync"
)

type OpenListenerFunc func(map[string]interface{}, string) (net.Listener, error)

var (
	listenerTypes   = make(map[string]OpenListenerFunc)
	listenerTypesMu sync.RWMutex
)

// RegisterListenerType makes a listener type available by the provided name.
// Use this function to extend go-listener with your own custom listener types.
//
// If RegisterListenerType is called twice with the same name or if
// openListener is nil, it panics.
func RegisterListenerType(name string, openListener OpenListenerFunc) {
	listenerTypesMu.Lock()
	defer listenerTypesMu.Unlock()

	if openListener == nil {
		panic("RegisterListenerType: openListener is nil")
	}
	if _, isDup := listenerTypes[name]; isDup {
		panic("RegisterListenerType: called twice for " + name)
	}
	listenerTypes[name] = openListener
}

func getOpenListenerFunc(listenerType string) OpenListenerFunc {
	listenerTypesMu.RLock()
	defer listenerTypesMu.RUnlock()
	return listenerTypes[listenerType]
}
