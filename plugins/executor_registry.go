package plugins

import (
	"errors"
	"goldenglow/node"
)

type executeRegistry struct {
	handlers map[string]ExecuteHandler
}

func (reg *executeRegistry) Register(name string, method ExecuteHandler) error {
	if name == "" {
		return errors.New("empty name")
	}
	if method == nil {
		return errors.New("nil method")
	}
	if reg.handlers == nil {
		reg.handlers = make(map[string]ExecuteHandler)
	}
	reg.handlers[name] = method
	return nil
}

func (reg *executeRegistry) Init() (string, node.Creator) {
	if reg.handlers == nil {
		reg.handlers = make(map[string]ExecuteHandler)
	}
	defaultCreator := func(b node.Base) node.Item {
		return &baseNode{
			handlers: reg.handlers,
		}
	}
	return KeyDefault, defaultCreator
}
func NewExecuteRegistry() ExecuteRegistry {
	return &executeRegistry{
		handlers: make(map[string]ExecuteHandler),
	}
}
