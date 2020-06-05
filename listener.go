package listeners // import "src.agwa.name/go-listeners"

import (
	"net"
	"sync"
)

type OpenListenerFunc func(map[string]interface{}, string) (net.Listener, error)

var (
	listenerTypes   = make(map[string]OpenListenerFunc)
	listenerTypesMu sync.Mutex
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
	listenerTypesMu.Lock()
	defer listenerTypesMu.Unlock()
	return listenerTypes[listenerType]
}
