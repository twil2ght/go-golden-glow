package workqueue

import "sync"

type Queue[T any] interface {
	Add(item T)
	Get() (item T, shutdown bool)
	Len() int
	Shutdown()
}

type DefaultQueue[T any] struct {
	items    []T
	cond     *sync.Cond
	shutdown bool
}

// New creates a new thread-safe queue
func New[T any]() Queue[T] {
	return &DefaultQueue[T]{
		items: make([]T, 0),
		cond:  sync.NewCond(&sync.Mutex{}),
	}
}

func (d *DefaultQueue[T]) Add(item T) {
	d.cond.L.Lock()
	defer d.cond.L.Unlock()

	if d.shutdown {
		return
	}

	d.items = append(d.items, item)
	d.cond.Signal()
}

func (d *DefaultQueue[T]) Get() (item T, shutdown bool) {
	d.cond.L.Lock()
	defer d.cond.L.Unlock()

	for len(d.items) == 0 && !d.shutdown {
		d.cond.Wait()
	}

	// Return zero value if shutdown and empty
	if len(d.items) == 0 {
		return *new(T), d.shutdown
	}

	item = d.items[0]
	d.items[0] = *new(T)
	d.items = d.items[1:]
	return item, d.shutdown
}

func (d *DefaultQueue[T]) Len() int {
	d.cond.L.Lock()
	defer d.cond.L.Unlock()
	return len(d.items)
}

// Shutdown unblocks all waiting Get() calls
func (d *DefaultQueue[T]) Shutdown() {
	d.cond.L.Lock()
	defer d.cond.L.Unlock()
	d.shutdown = true
	d.cond.Broadcast()
}
