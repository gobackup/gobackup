package notifier

import (
	"fmt"
	"sync"
)

// Factory creates a Notifier instance from Base config
// Returns (Notifier, error) to support notifiers that may fail during initialization
type Factory func(base *Base) (Notifier, error)

// Registry holds all registered notifier factories
type Registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

// NewRegistry creates a new empty registry
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]Factory),
	}
}

// Register adds a factory for the given notifier type
func (r *Registry) Register(name string, factory Factory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

// Get returns the factory for the given notifier type
func (r *Registry) Get(name string) (Factory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if factory, ok := r.factories[name]; ok {
		return factory, nil
	}
	return nil, fmt.Errorf("unsupported notifier type: %s", name)
}

// Has checks if a notifier type is registered
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.factories[name]
	return ok
}

// Types returns all registered notifier type names
func (r *Registry) Types() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]string, 0, len(r.factories))
	for name := range r.factories {
		types = append(types, name)
	}
	return types
}

// DefaultRegistry is the global registry used by Run()
var DefaultRegistry = NewRegistry()

// Register is a convenience function for DefaultRegistry.Register
func Register(name string, factory Factory) {
	DefaultRegistry.Register(name, factory)
}
