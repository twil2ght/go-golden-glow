package node

import (
	"fmt"
	"goldenglow/variable"
	"strings"
)

type Creator func(base Base) Item

type Checkable interface {
	Check() error
}
type Extractable interface {
	Extract() (variable.Item, error)
}
type Encoder interface {
	Do(nodeValue string) string
	Match(nv1, nv2 string) bool
}

type Regulator interface {
	Do(nodeValue string) string
}

type Registry interface {
	Register(name string, method Creator) error
}
type Pool interface {
	Add(node Item)
}
type Factory interface {
	New(nodeValue string) (Item, error)
	Registry
	Pool
}
type factory struct {
	parser     variable.Parser
	regulator  Regulator
	pool       Set
	registries map[string]Creator
}

func NewFactory(p variable.Parser, r Regulator) (Factory, error) {
	if r == nil {
		return nil, fmt.Errorf("factory init: Regulator is nil")
	}
	if p == nil {
		return nil, fmt.Errorf("factory init: parser is nil")
	}
	return &factory{
		parser:     p,
		regulator:  r,
		pool:       make(Set),
		registries: make(map[string]Creator),
	}, nil
}
func (f *factory) New(val string) (Item, error) {
	if val == "" {
		return nil, fmt.Errorf("node val can't be empty")
	}
	base := Base{
		val:       val,
		variables: make(variable.Set),
		parser:    f.parser,
		regulator: f.regulator,
	}
	var (
		node    Item = &base
		warning string
	)
	event := head(val)
	if creator, ok := f.registries[event]; ok {
		node = creator(base)
	} else {
		warning = fmt.Sprintf("node %s :s unknown event '%s' (use default node)", val, event)
	}

	f.Add(node)

	if warning == "" {
		return node, nil
	}
	return node, fmt.Errorf("%s", warning)
}
func (f *factory) Register(event string, creator Creator) error {
	head := "node registry"
	if event == "" {
		return fmt.Errorf("%s: event can't be empty", head)
	}
	if creator == nil {
		return fmt.Errorf("%s: creator can't be nil", head)
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
