package executor

import (
	"errors"
	"goldenglow/node"
	"goldenglow/pkg/log"
	"goldenglow/utils"
)

var (
	logger = log.Default()
)

type executeRegistry struct {
	handlers map[string]Handler
	nodeReg  node.Registry
}

func (reg *executeRegistry) Register(pluginName string, method Handler) error {
	if err := utils.NotNull(
		"pluginName", pluginName,
		"method", method,
	); err != nil {
		return err
	}
	if reg.handlers == nil {
		reg.handlers = make(map[string]Handler)
	}
	reg.handlers[pluginName] = method
	logger.Info("executor:register successfully", "plugin", pluginName)
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
		return &BaseNode{
			handlers: reg.handlers,
		}
	}
	logger.Info("executor:register to node registry", "handler_amount", len(reg.handlers))
	return nReg.Register(KeyDefault, defaultCreator)
}
func NewRegistry(nodeReg node.Registry) Registry {
	return &executeRegistry{
		handlers: make(map[string]Handler),
		nodeReg:  nodeReg,
	}
}
