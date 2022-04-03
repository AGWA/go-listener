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
