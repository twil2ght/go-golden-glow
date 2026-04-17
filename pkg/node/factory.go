package node

import (
	"goldenglow/pkg/registry"
	"strings"
	"sync"
)

type Creator func(value string) Interface
type Factory interface {
	Create(value string) Interface
	CreatorRegistry() registry.Interface[Creator]
}
type factory struct {
	creators registry.Interface[Creator]
	items    registry.Interface[Interface]
	mu       sync.RWMutex
}

func (f *factory) CreatorRegistry() registry.Interface[Creator] {
	return f.creators
}

func (f *factory) Create(value string) Interface {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if item, err := f.items.Get(value); err == nil {
		return item
	}
	if creator, err := f.creators.Get(GetHead(value)); err == nil {
		item := creator(value)
		f.items.Register(value, item)
		return item
	}
	return New(value)
}

func NewFactory() Factory {
	return &factory{
		creators: registry.New[Creator](),
		items:    registry.New[Interface](),
	}
}
func DefaultFactory() Factory {
	return NewFactory()
}
func GetHead(str string) string {
	fields := strings.Fields(str)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
