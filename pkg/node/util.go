package node

import (
	"goldenglow/variable"
	"sort"
	"strings"
)

func GenVariableState(vSet variable.Set) string {
	// Extract keys from the variable set
	keys := make([]string, 0, len(vSet))
	for key := range vSet {
		keys = append(keys, key)
	}

	// Sort keys to ensure consistent ordering
	sort.Strings(keys)

	// Combine each key with its value into a single string
	var parts []string
	for _, key := range keys {
		if item, ok := vSet[key]; ok {
			parts = append(parts, key+item.Value())
		}
	}

	// Join all parts together
	return strings.Join(parts, "")
}
