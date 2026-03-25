package checker

import (
	"fmt"
	"goldenglow/executor"
)

const (
	KeyChecker = "[node:checker]"
)

type baseChecker struct {
	executor.BaseNode
	handlers map[string]Handler
}

func (b *baseChecker) Check() error {
	params := b.GetParams()
	pluginID := params[executor.KeyNamespace]

	if pluginID == "" {
		return fmt.Errorf("pluginID(%s) not found in params", executor.KeyNamespace)
	}

	handler := b.handlers[pluginID]
	if handler == nil {
		return fmt.Errorf("plugin %s not registered", pluginID)
	}

	b.SetState(handler(params))
	return nil
}
