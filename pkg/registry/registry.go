package registry

import "sync"

type Registry[T any] interface {
	Register(name string, value T)
	Unregister(name string)
	Get(name string) (T, error)
	Range(f func(T) bool)
	Len() int
	Keys() []string
}
type DefaultRegistry[T any] struct {
	items map[string]T
	keys  []string
	mu    sync.RWMutex
}

func (d *DefaultRegistry[T]) Unregister(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.items, name)
	for i, k := range d.keys {
		if k == name {
			d.keys = append(d.keys[:i], d.keys[i+1:]...)
			break
		}
	}
}

func (d *DefaultRegistry[T]) Register(name string, value T) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.items[name]; exists {
		return
	}
	d.items[name] = value
	d.keys = append(d.keys, name)
}

func (d *DefaultRegistry[T]) Get(name string) (T, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.items[name], nil
}

func (d *DefaultRegistry[T]) Range(f func(T) bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	for _, k := range d.keys {
		if !f(d.items[k]) {
			return
		}
	}
	return
}

func (d *DefaultRegistry[T]) Len() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.items)
}

func (d *DefaultRegistry[T]) Keys() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.keys
}

// New creates a new DefaultRegistry instance.
func New[T any]() Registry[T] {
	return &DefaultRegistry[T]{
		items: make(map[string]T),
		keys:  make([]string, 0),
	}
}
