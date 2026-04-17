package handler

import (
	"goldenglow/pkg/registry"
	"strings"
)

var (
	KeyDist       = "dist"
	KeyNamespace  = "namespace"
	NodeExecutor  = "[node:executor]"
	NodeChecker   = "[node:checker]"
	NodeExtractor = "[node:extractor]"
)

type Parameters registry.Interface[string]

func GetParameters(nodeValueWithNoVar string) Parameters {
	params := registry.New[string]()

	if nodeValueWithNoVar == "" {
		return params
	}

	// Parse bracket pairs with nesting support
	i := 0
	for i < len(nodeValueWithNoVar) {
		// Find opening bracket
		if nodeValueWithNoVar[i] != '[' {
			i++
			continue
		}

		start := i
		i++

		// Find the colon separator (not inside nested brackets)
		colonIdx := -1
		depth := 0
		for i < len(nodeValueWithNoVar) {
			ch := nodeValueWithNoVar[i]
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
		if i >= len(nodeValueWithNoVar) || nodeValueWithNoVar[i] != ']' || colonIdx == -1 {
			continue
		}

		// Extract key and value
		key := strings.TrimSpace(nodeValueWithNoVar[start+1 : colonIdx])
		value := strings.TrimSpace(nodeValueWithNoVar[colonIdx+1 : i])

		if key != "" {
			params.Register(key, value)
		}

		i++
	}

	return params
}
