package support

import (
	"sync"
)

var registry = struct {
	sync.RWMutex
	native map[string]any
}{
	native: make(map[string]any),
}

func Register(key string, val any) {
	registry.Lock()
	registry.native[key] = val
	registry.Unlock()
}

func GetValue(key string) (any, bool) {
	registry.RLock()
	defer registry.RUnlock()
	module, ok := registry.native[key]
	return module, ok
}

func RemoveValue(name string) {
	registry.Lock()
	delete(registry.native, name)
	registry.Unlock()
}
