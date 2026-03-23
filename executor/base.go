package executor

import (
	"fmt"
	"goldenglow/node"
	"regexp"
	"strings"
)

type baseNode struct {
	node.Base
	handlers map[string]Handler
}

const (
	KeyNamespace = "namespace"
	KeyDefault   = "[node]"
)

func (d *baseNode) Execute() error {
	params := d.GetParams()
	pluginID := params[KeyNamespace]

	if pluginID == "" {
		return fmt.Errorf("pluginID(%s) not found in params", KeyNamespace)
	}

	handler := d.handlers[pluginID]
	if handler == nil {
		return fmt.Errorf("plugin %s not registered", pluginID)
	}

	return handler(params)
}

var kvRegex = regexp.MustCompile(`\[([^:\]]+):([^]]+)]`)

// GetParams 正则解析 [key:value] 参数
func (d *baseNode) GetParams() map[string]string {
	params := make(map[string]string)

	nodeValue, _ := d.ToText()
	if nodeValue == "" {
		return params
	}

	// 匹配所有 [k:v]
	matches := kvRegex.FindAllStringSubmatch(nodeValue, -1)
	for _, match := range matches {
		// match[1] = key
		// match[2] = value
		key := strings.TrimSpace(match[1])
		value := strings.TrimSpace(match[2])
		params[key] = value
	}

	return params
}
