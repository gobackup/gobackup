package database

import (
	"fmt"
	"sync"
)

// Factory creates a Database instance from a Base configuration.
// Each database type should register its factory using Register().
type Factory func(base Base) Database

var (
	registry   = make(map[string]Factory)
	registryMu sync.RWMutex
)

// Register adds a database factory to the registry.
// This should be called in the init() function of each database type.
// Example:
//
//	func init() {
//	    Register("mysql", func(base Base) Database {
//	        return &MySQL{Base: base}
//	    })
//	}
func Register(name string, factory Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if factory == nil {
		panic(fmt.Sprintf("database: Register factory is nil for %s", name))
	}
	if _, dup := registry[name]; dup {
		panic(fmt.Sprintf("database: Register called twice for %s", name))
	}
	registry[name] = factory
}

// Get returns the factory for the given database type.
// Returns nil if the type is not registered.
func Get(name string) Factory {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[name]
}

// ListTypes returns all registered database type names.
func ListTypes() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
