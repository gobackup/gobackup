package storage

import (
	"fmt"
	"sync"
)

// Factory creates a Storage instance from a Base configuration.
// Each storage type should register its factory using Register().
type Factory func(base Base) Storage

var (
	registry   = make(map[string]Factory)
	registryMu sync.RWMutex
)

// Register adds a storage factory to the registry.
// This should be called in the init() function of each storage type.
// Example:
//
//	func init() {
//	    Register("s3", func(base Base) Storage {
//	        return &S3{Base: base}
//	    })
//	}
func Register(name string, factory Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if factory == nil {
		panic(fmt.Sprintf("storage: Register factory is nil for %s", name))
	}
	if _, dup := registry[name]; dup {
		panic(fmt.Sprintf("storage: Register called twice for %s", name))
	}
	registry[name] = factory
}

// Get returns the factory for the given storage type.
// Returns nil if the type is not registered.
func Get(name string) Factory {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[name]
}

// ListTypes returns all registered storage type names.
func ListTypes() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
