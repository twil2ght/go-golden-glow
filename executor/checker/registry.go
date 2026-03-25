package checker

import (
	"errors"
	"goldenglow/node"
	"goldenglow/pkg/log"
	"goldenglow/utils"
)

var (
	logger = log.Default()
)

type registry struct {
	handlers map[string]Handler
	nodeReg  node.Registry
}

func (reg *registry) Register(pluginName string, method Handler) error {
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
	logger.Info("checker:register successfully", "plugin", pluginName)
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
	logger.Info("checker:register to node registry", "handler_amount", len(reg.handlers))
	return nReg.Register(KeyChecker, defaultCreator)
}
func NewRegistry(nodeReg node.Registry) Registry {
	return &registry{
		handlers: make(map[string]Handler),
		nodeReg:  nodeReg,
	}
}
