package executor

import (
	"errors"
	"goldenglow/node"
)

type executeRegistry struct {
	handlers map[string]Handler
	nodeReg  node.Registry
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

func (reg *executeRegistry) RunAll() error {
	if reg.handlers == nil {
		return errors.New("empty handler")
	}
	return reg.OnRegisterNodeRegistry(reg.nodeReg)
}
func (reg *executeRegistry) OnRegisterNodeRegistry(nReg node.Registry) error {
	defaultCreator := func(b node.Base) node.Item {
		return &baseNode{
			handlers: reg.handlers,
		}
	}
	return nReg.Register(KeyDefault, defaultCreator)
}
func NewRegistry(nodeReg node.Registry) Registry {
	return &executeRegistry{
		handlers: make(map[string]Handler),
		nodeReg:  nodeReg,
	}
}
