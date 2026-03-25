package checker

import (
	"errors"
	"goldenglow/node"
)

type registry struct {
	handlers map[string]Handler
	nodeReg  node.Registry
}

func (reg *registry) Register(name string, method Handler) error {
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

func (reg *registry) RunAll() error {
	if reg.handlers == nil {
		return errors.New("empty handler")
	}
	return reg.OnRegisterNodeRegistry(reg.nodeReg)
}
func (reg *registry) OnRegisterNodeRegistry(nReg node.Registry) error {
	defaultCreator := func(b node.Base) node.Item {
		return &baseChecker{
			handlers: reg.handlers,
		}
	}
	return nReg.Register(KeyChecker, defaultCreator)
}
func NewRegistry(nodeReg node.Registry) Registry {
	return &registry{
		handlers: make(map[string]Handler),
		nodeReg:  nodeReg,
	}
}
