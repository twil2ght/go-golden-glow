package node

import (
	"fmt"
	"goldenglow/utils"
	"goldenglow/variable"
	"strings"
)

type factory struct {
	parser     variable.Parser
	pool       Set
	registries map[string]Creator
}

func NewFactory(p variable.Parser) (Factory, error) {
	if err := utils.NotNull("parser", p); err != nil {
		return nil, fmt.Errorf("node factory init:%w", err)
	}
	return &factory{
		parser:     p,
		pool:       make(Set),
		registries: make(map[string]Creator),
	}, nil
}
func (f *factory) NewFromPool(val string) (Item, error) {
	if item, ok := f.pool[val]; ok {
		return item, nil
	}
	return f.New(val)
}
func (f *factory) New(val string) (Item, error) {
	if val == "" {
		return nil, fmt.Errorf("node val can't be empty")
	}
	base := Base{
		val:            val,
		variables:      make(variable.Set),
		parser:         f.parser,
		variableState:  make(map[string]map[string]bool),
		variableSetHub: make(map[string]variable.Set),
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
	head := "node registry"
	if err := utils.NotNull("creator", creator, "event", event); err != nil {
		return fmt.Errorf("%s: %w", head, err)
	}
	f.registries[event] = creator
	return nil
}
func (f *factory) Add(n Item) {
	if _, ok := f.pool[n.Value()]; !ok {
		f.pool[n.Value()] = n
	}
}

func head(str string) string {
	fields := strings.Fields(str)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
