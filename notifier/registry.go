package notifier

import (
	"fmt"
	"sync"
)

// Factory creates a Notifier instance from a Base configuration.
// Each notifier type should register its factory using Register().
type Factory func(base *Base) (Notifier, error)

var (
	registry   = make(map[string]Factory)
	registryMu sync.RWMutex
)

// Register adds a notifier factory to the registry.
// This should be called in the init() function of each notifier type.
// Example:
//
//	func init() {
//	    Register("slack", func(base *Base) (Notifier, error) {
//	        return NewSlack(base), nil
//	    })
//	}
func Register(name string, factory Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if factory == nil {
		panic(fmt.Sprintf("notifier: Register factory is nil for %s", name))
	}
	if _, dup := registry[name]; dup {
		panic(fmt.Sprintf("notifier: Register called twice for %s", name))
	}
	registry[name] = factory
}

// Get returns the factory for the given notifier type.
// Returns nil if the type is not registered.
func Get(name string) Factory {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[name]
}

// ListTypes returns all registered notifier type names.
func ListTypes() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
