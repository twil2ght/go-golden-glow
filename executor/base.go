package executor

import (
	"fmt"
	"goldenglow/node"
	"strings"
)

type BaseNode struct {
	node.Base
	handlers map[string]Handler
}

const (
	KeyNamespace = "namespace"
	KeyDefault   = "[node]"
)

func (d *BaseNode) Execute() error {
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

// GetParams parses [key:value] parameters, supporting nested brackets in values
func (d *BaseNode) GetParams() map[string]string {
	params := make(map[string]string)

	nodeValue, _ := d.ToText()
	if nodeValue == "" {
		return params
	}

	// Parse bracket pairs with nesting support
	i := 0
	for i < len(nodeValue) {
		// Find opening bracket
		if nodeValue[i] != '[' {
			i++
			continue
		}

		start := i
		i++

		// Find the colon separator (not inside nested brackets)
		colonIdx := -1
		depth := 0
		for i < len(nodeValue) {
			ch := nodeValue[i]
			if ch == '[' {
				depth++
			} else if ch == ']' {
				if depth == 0 {
					break
				}
				depth--
			} else if ch == ':' && depth == 0 && colonIdx == -1 {
				colonIdx = i
			}
			i++
		}

		// Check if we found a valid closing bracket and colon
		if i >= len(nodeValue) || nodeValue[i] != ']' || colonIdx == -1 {
			continue
		}

		// Extract key and value
		key := strings.TrimSpace(nodeValue[start+1 : colonIdx])
		value := strings.TrimSpace(nodeValue[colonIdx+1 : i])

		if key != "" {
			params[key] = value
		}

		i++
	}

	return params
}
