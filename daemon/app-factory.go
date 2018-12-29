package daemon

import (
	"sync"
)

var (
	mu        sync.RWMutex
	factories = make(map[string]DaemonApplication)
)

func RegisterDaemonApp(pkgName string, f DaemonApplication) {
	mu.Lock()
	defer mu.Unlock()

	if f == nil {
		panic("AppFactory is nil")
	}
	if _, exist := factories[pkgName]; exist {
		panic("AppFactory already registered")
	}

	factories[pkgName] = f
}

func NewDaemonApp(appName string) (DaemonApplication, bool) {
	mu.RLock()
	defer mu.RUnlock()
	if f, exist := factories[appName]; exist {
		return f.NewInstance(), true
	}
	return nil, false
}
