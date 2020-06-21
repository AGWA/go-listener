package cert

import (
	"sync"
	"time"
)

var (
	globalFileCache     *FileCache
	globalFileCacheOnce sync.Once
)

func cleanGlobalFileCache() {
	for range time.Tick(24 * time.Hour) {
		globalFileCache.Clean()
	}
}

func GlobalFileCache() *FileCache {
	globalFileCacheOnce.Do(func() {
		globalFileCache = NewFileCache()
		go cleanGlobalFileCache()
	})
	return globalFileCache
}
