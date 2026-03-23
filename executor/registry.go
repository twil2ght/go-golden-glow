package executor

import (
	"errors"
	"goldenglow/node"
)

type executeRegistry struct {
	handlers map[string]Handler
}

func (reg *executeRegistry) Register(name string, method Handler) error {
	if name == "" {
		return errors.New("empty name")
	}
	if method == nil {
		return errors.New("nil method")
	}
	if reg.handlers == nil {
		reg.handlers = make(map[string]Handler)
	}
	reg.handlers[name] = method
	return nil
}

func (reg *executeRegistry) RunAll() (string, node.Creator) {
	if reg.handlers == nil {
		reg.handlers = make(map[string]Handler)
	}
	defaultCreator := func(b node.Base) node.Item {
		return &baseNode{
			handlers: reg.handlers,
		}
	}
	return KeyDefault, defaultCreator
}
func NewExecuteRegistry() Registry {
	return &executeRegistry{
		handlers: make(map[string]Handler),
	}
}
