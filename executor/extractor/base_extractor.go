package extractor

import (
	"fmt"
	"goldenglow/executor"
	"goldenglow/variable"
)

const (
	KeyExtractor = "[node:extractor]"
	KeyDist      = "dist"
)

type baseExtractor struct {
	executor.BaseNode
	handlers map[string]Handler
}

func (b *baseExtractor) Extract() (variable.Item, error) {
	params := b.GetParams()
	pluginID := params[executor.KeyNamespace]

	if pluginID == "" {
		return nil, fmt.Errorf("pluginID(%s) not found in params", executor.KeyNamespace)
	}

	handler := b.handlers[pluginID]
	if handler == nil {
		return nil, fmt.Errorf("plugin %s not registered", pluginID)
	}

	return handler(params)
}
func (b *baseExtractor) SetState(_ bool) {
	b.Base.SetState(true)
}
func (b *baseExtractor) OK() bool {
	return true
}

// ExtractTarget returns the variable name that will receive the extracted value
func (b *baseExtractor) ExtractTarget() string {
	params := b.GetParams()
	return params[KeyDist]
}
