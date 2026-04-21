package container

import (
	"goldenglow/m"
	"goldenglow/pkg/node"
	"goldenglow/variable"
)

// getCompatibleSets
func getCompatibleSets(n node.Interface, base variable.Set) ([]variable.Set, bool) {
	if len(n.VarKeys()) == 0 {
		return nil, true
	}
	registryN := n.VarSetRegistry()
	if registryN.Len() == 0 {
		return nil, false
	}

	var res []variable.Set
	for _, s := range registryN.Keys() {
		vars, _ := registryN.Get(s)
		if isCompatible(vars, base) {
			res = append(res, vars)
		}
	}
	if len(res) == 0 {
		return nil, false
	}
	return res, true
}

// isCompatible check if a set is compatible with the base set
func isCompatible(set, base variable.Set) bool {
	for k, vBase := range base {
		if v, ok := set[k]; ok && v.Value() != vBase.Value() {
			return false
		}
	}
	return true
}

// mergeTwoVarSets
func mergeTwoVarSets(a, b variable.Set) (variable.Set, bool) {
	merged := make(variable.Set)
	for k, v := range a {
		merged[k] = v.Copy()
	}
	for k, v := range b {
		if exist, ok := merged[k]; ok {
			if exist.Value() != v.Value() {
				return nil, false
			}
			continue
		}
		merged[k] = v.Copy()
	}
	return merged, true
}

func mergeVariables(
	excludeNode node.Interface,
	ts m.Map[node.Interface],
	varSet variable.Set,
) bool {
	var nodes []node.Interface
	for _, t := range ts {
		if t.Value() != excludeNode.Value() {
			nodes = append(nodes, t)
		}
	}
	if len(nodes) == 0 {
		return true
	}

	baseVars := variable.Copy(varSet)

	startNode, startSets, valid := findStart(nodes, baseVars)
	if startNode == nil || len(startSets) == 0 {
		// no node has varSet ->legal
		if !valid {
			return false
		}
		return true
	}

	for _, currentNode := range nodes {
		if currentNode.Value() == startNode.Value() {
			continue
		}

		cache := make([]variable.Set, 0, len(startSets))
		for _, s := range startSets {
			compatibleSets, ok := getCompatibleSets(currentNode, s)
			//has varSet but incompatible
			if !ok {
				continue
			}
			//no varSet but this is legal
			if len(compatibleSets) == 0 {
				cache = append(cache, s)
			}

			for _, vSet := range compatibleSets {
				merged, ok2 := mergeTwoVarSets(s, vSet)
				if ok2 {
					cache = append(cache, merged)
				}
			}
		}

		if len(cache) == 0 {
			break
		}

		startSets = cache
	}

	if len(startSets) > 0 {
		for k, v := range startSets[0] {
			if _, exists := varSet[k]; !exists {
				varSet[k] = v
			}
		}
		return true
	}
	return false
}

// findStart finds the first node that has varSet if it exists
func findStart(nodes []node.Interface, baseVars variable.Set) (node.Interface, []variable.Set, bool) {
	var hasValidItem bool
	for _, n := range nodes {
		sets, valid := getCompatibleSets(n, baseVars)
		if valid {
			hasValidItem = true
			if len(sets) > 0 {
				return n, sets, true
			}
		}
	}
	return nil, nil, hasValidItem
}
