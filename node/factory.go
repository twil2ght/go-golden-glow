package node

import (
	"fmt"
	"goldenglow/m"
	"goldenglow/utils"
	"goldenglow/variable"
	"strings"
	"sync"
)

type factory struct {
	parser     variable.Parser
	pool       Set
	registries map[string]Creator
	mu         *sync.RWMutex
}

func NewFactory(p variable.Parser) (Factory, error) {
	if err := utils.NotNull("parser", p); err != nil {
		return nil, fmt.Errorf("node factory init:%w", err)
	}
	return &factory{
		parser:     p,
		pool:       make(Set),
		registries: make(map[string]Creator),
		mu:         &sync.RWMutex{},
	}, nil
}
func (f *factory) NewFromPool(val string) (Item, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if item, ok := f.pool[val]; ok {
		return item, nil
	}
	return f.New(val)
}
func (f *factory) New(val string) (Item, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if val == "" {
		return nil, fmt.Errorf("node val can't be empty")
	}
	base := Base{
		val:            val,
		variables:      make(variable.Set),
		parser:         f.parser,
		variableState:  make(m.Hash),
		variableSetHub: make(map[string]variable.Set),
		mu:             &sync.RWMutex{},
	}
	var (
		node Item = &base
	)
	event := head(val)
	if creator, ok := f.registries[event]; ok {
		node = creator(base)
	}

	f.Add(node)
	return node, nil
}
func (f *factory) Register(event string, creator Creator) error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	head := "node registry"
	if err := utils.NotNull("creator", creator, "event", event); err != nil {
		return fmt.Errorf("%s: %w", head, err)
	}
	f.registries[event] = creator
	return nil
}
func (f *factory) Add(n Item) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if _, ok := f.pool[n.Value()]; !ok {
		f.pool[n.Value()] = n
	}
}

func (f *factory) ResetPool() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pool = make(Set)
	//for _, item := range f.pool {
	//	item.Reset()
	//}
}

func head(str string) string {
	fields := strings.Fields(str)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
