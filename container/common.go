package container

import (
	"errors"
	"fmt"
	"goldenglow/node"
	"goldenglow/pkg/log"
	"goldenglow/variable"
)

// 处理单个变量匹配项
func variableGen(
	index int,
	dMatch []string,
	tMatches [][]string,
	GivenSet variable.Set,
	targetVars variable.Set,
) error {
	// 目标变量名
	token := dMatch[0]
	// 源变量名
	varKey := tMatches[index][0]
	prevVar := GivenSet[varKey]
	logger.Debug("prevVar", "", prevVar)
	if prevVar == nil {
		return log.NotFound("variable:" + varKey)
	}

	// 生成新变量
	targetVars[token] = variable.New(token, prevVar.Value())
	return nil
}
func (b *Base) variableGen(T, dist node.Item, GivenSet variable.Set) (variable.Set, error) {
	tMatches := b.varReg.FindAllStringSubmatch(T.Value(), -1)
	dMatches := b.varReg.FindAllStringSubmatch(dist.Value(), -1)
	if len(tMatches) != len(dMatches) {
		return nil, fmt.Errorf("variable count mismatch: %d != %d", len(tMatches), len(dMatches))
	}

	variables := make(variable.Set)
	for i, dMatch := range dMatches {
		if err := variableGen(i, dMatch, tMatches, GivenSet, variables); err != nil {
			return nil, fmt.Errorf("variableGen:%s", err)
		}
	}
	if len(dMatches) != len(variables) {
		return nil, fmt.Errorf("variableGen: variable count mismatch: %d != %d", len(dMatches), len(variables))
	}
	return variables, nil
}

func (b *Base) mergeVariables(excludeNode node.Item) error {
	// Collect all trigger nodes except the excluded one
	var otherNodes []node.Item
	for _, t := range b.tNodes {
		if t.Value() == excludeNode.Value() {
			continue
		}
		otherNodes = append(otherNodes, t)
	}

	// Try to find a unified variable set that is compatible with current container variables
	// and satisfies all other trigger nodes
	unifiedVars := b.findCompatibleVariableSet(otherNodes)

	// Merge the unified variables into the container's variables
	for name, varItem := range unifiedVars {
		if _, exists := b.variables[name]; !exists {
			b.variables[name] = varItem
		}
	}
	return nil
}

// findCompatibleVariableSet finds a set of variables that is compatible with the current
// container variables and can satisfy all other trigger nodes
func (b *Base) findCompatibleVariableSet(nodes []node.Item) variable.Set {
	if len(nodes) == 0 {
		return nil
	}

	// Start with current container variables as the base
	baseVars := variable.Copy(b.variables)

	// Start with the first node's states that are compatible with baseVars
	candidates := b.getCompatibleVariableSets(nodes[0], baseVars)

	// Iteratively refine candidates with each subsequent node
	for i := 1; i < len(nodes); i++ {
		nodeStates := b.getCompatibleVariableSets(nodes[i], baseVars)
		candidates = b.mergeCompatibleStates(candidates, nodeStates)
		if len(candidates) == 0 {
			break
		}
	}

	// Return the first valid unified set, or fallback to simple merge
	if len(candidates) > 0 {
		return candidates[0]
	}

	// Fallback: return merged variables from current node states
	return b.fallbackMerge(nodes)
}

// getCompatibleVariableSets returns variable sets from a node that are compatible with baseVars
// (i.e., they have the same values for any shared variables)
func (b *Base) getCompatibleVariableSets(n node.Item, baseVars variable.Set) []variable.Set {
	hub := n.VarSetRegistry()
	if len(hub) == 0 {
		// Fallback to current variables if hub is empty
		vars := n.Vars()
		if len(vars) > 0 && b.isCompatibleWithBase(vars, baseVars) {
			return []variable.Set{vars}
		}
		return nil
	}

	var compatibleSets []variable.Set
	for _, vSet := range hub {
		if b.isCompatibleWithBase(vSet, baseVars) {
			compatibleSets = append(compatibleSets, vSet)
		}
	}
	return compatibleSets
}

// isCompatibleWithBase checks if a variable set is compatible with baseVars
// (they have the same values for any shared variables)
func (b *Base) isCompatibleWithBase(vSet, baseVars variable.Set) bool {
	for key, baseItem := range baseVars {
		if item, ok := vSet[key]; ok {
			if item.Value() != baseItem.Value() {
				return false // Different values for shared variable
			}
		}
	}
	return true
}

// mergeCompatibleStates combines two lists of variable sets, keeping only compatible combinations
// Two states are compatible if they have the same values for shared variables
func (b *Base) mergeCompatibleStates(setA, setB []variable.Set) []variable.Set {
	var result []variable.Set

	for _, a := range setA {
		for _, b_ := range setB {
			if merged, ok := b.mergeTwoStates(a, b_); ok {
				result = append(result, merged)
			}
		}
	}

	return result
}

// mergeTwoStates attempts to merge two variable sets
// Returns (mergedSet, true) if compatible, (nil, false) otherwise
func (b *Base) mergeTwoStates(setA, setB variable.Set) (variable.Set, bool) {
	merged := make(variable.Set)

	// Copy all from setA
	for k, v := range setA {
		merged[k] = v.Copy()
	}

	// Merge from setB, checking compatibility
	for k, v := range setB {
		if existing, exists := merged[k]; exists {
			// Check if values match for shared variables
			if existing.Value() != v.Value() {
				return nil, false // Incompatible
			}
			// Same value, skip
			continue
		}
		merged[k] = v.Copy()
	}

	return merged, true
}

// fallbackMerge performs a simple merge of current variables from all nodes
func (b *Base) fallbackMerge(nodes []node.Item) variable.Set {
	merged := make(variable.Set)
	for _, n := range nodes {
		for name, varItem := range n.Vars() {
			if _, exists := merged[name]; !exists {
				merged[name] = varItem.Copy()
			}
		}
	}
	return merged
}

func (b *Base) processNode(t node.Item) error {
	// Handle Checkable nodes
	if checker, ok := t.(node.Checkable); ok {
		return b.handleChecker(checker)
	}

	// Handle Extractable nodes
	if extractor, ok := t.(node.Extractable); ok {
		return b.handleExtractor(extractor)
	}

	return nil
}

func (b *Base) handleChecker(checker node.Checkable) error {
	t := checker.(node.Item)
	finalVars, err := b.prepareVariables(t, "")
	if err != nil {
		return err
	}

	if err := t.SetAndRegisterVars(finalVars); err != nil {
		return fmt.Errorf("set variable failed: %w", err)
	}

	if err := checker.Check(); err != nil {
		return fmt.Errorf("check failed: %w", err)
	}

	return nil
}

func (b *Base) handleExtractor(extractor node.Extractable) error {
	t := extractor.(node.Item)
	extractTarget := extractor.ExtractTarget()
	finalVars, err := b.prepareVariables(t, extractTarget)
	if err != nil {
		return err
	}

	if err := t.SetAndRegisterVars(finalVars); err != nil {
		return fmt.Errorf("set variable failed: %w", err)
	}

	varb, err := extractor.Extract()
	if err != nil {
		return fmt.Errorf("extract failed: %w", err)
	}

	b.variables[varb.Name()] = varb
	if valueMap, ok := varb.(variable.ValueMap); ok {
		b.valueHash = valueMap.Get()
		b.valueHashKey = varb.Name()
		logger.Debug("extractor returned stateHub", "size", len(b.valueHash))
	}

	return nil
}

func (b *Base) prepareVariables(t node.Item, skipKey string) (variable.Set, error) {
	keys := t.VarKeys()
	finalVars := make(variable.Set, len(keys))
	var errs []error

	for _, key := range keys {
		if key == skipKey {
			continue
		}

		varb := b.variables[key]
		if varb == nil {
			errs = append(errs, fmt.Errorf("variable:%s not found", key))
			b.logAvailableVariables()
			continue
		}
		finalVars[key] = varb.Copy()
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return finalVars, nil
}

func (b *Base) logAvailableVariables() {
	for k := range b.variables {
		logger.Error("next: variable", "name", k, "value", b.variables[k].Value())
	}
}

// checkVariableStateMap verifies that the current variable values exist in the trigger's variableStateMap
func (b *Base) checkVariableStateMap(t node.Item) (bool, error) {
	stateMap := t.VarStateRegistry()
	logger.Debug("checkVariableStateMap", "node", t.Value(), "variableStateMap", stateMap)
	var variableSet = make(variable.Set)
	for _, key := range t.VarKeys() {
		if item, ok := b.variables[key]; ok {
			variableSet[key] = item
		} else {
			return false, fmt.Errorf("variable:%s not found", key)
		}
	}
	var state = node.GenVariableState(variableSet)
	if _, ok := t.VarStateRegistry()[state]; !ok {
		return false, fmt.Errorf("variable state:%s not found", state)
	}
	return true, nil
}

func (b *Base) findT(target node.Item, set node.Set) (node.Item, error) {
	var tar = b.encoder.Do(target.Value())
	for _, n := range set {
		if b.encoder.Do(n.Value()) == tar {
			return n, nil
		}
	}
	return nil, log.NotFound("external node")
}
