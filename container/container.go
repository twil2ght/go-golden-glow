package container

import (
	"errors"
	"fmt"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/pkg/log"
	"goldenglow/utils"
	"goldenglow/variable"
	"regexp"
)

var logger = log.Default()

type Item interface {
	ID() string
	TNode() node.Set
	RNode() node.Set
	Variables() variable.Set
	Do(external node.Item) (bool, error)
}
type Base struct {
	hashValue    string
	tNodes       node.Set
	rNodes       node.Set
	variables    variable.Set
	fetcher      Fetcher
	encoder      node.Encoder
	varReg       *regexp.Regexp
	valueHash    m.Hash
	valueHashKey string
}

func (b *Base) ID() string {
	return b.hashValue
}
func (b *Base) TNode() node.Set {
	return b.tNodes
}
func (b *Base) RNode() node.Set {
	return b.rNodes
}
func (b *Base) Variables() variable.Set {
	return b.variables
}
func (b *Base) Do(T node.Item) (bool, error) {
	err := b.setNode()
	if err != nil {
		logger.Debug("setNode failed", "id", b.ID(), "error", err)
		return false, err
	}

	err = b.ParseTrigger(T)
	if err != nil {
		logger.Debug("parseTrigger failed", "id", b.ID(), "error", err)
		return false, err
	}

	err = b.CheckAndExtract()
	if err != nil {
		logger.Debug("checkAndExtract failed", "id", b.ID(), "error", err)
		return false, err
	}

	ok, err := b.OK()
	if !ok {
		logger.Debug("container is not ok", "id", b.ID())
		if err != nil {
			return false, err
		}
		return false, fmt.Errorf("OK() failed")
	}

	err = b.Next()
	if err != nil {
		logger.Debug("next failed", "id", b.ID(), "error", err)
		return false, err
	}
	logger.Debug("container is ok", "id", b.ID())
	return true, nil
}
func (b *Base) setNode() error {
	head := "[setNode]"
	tn, err := b.fetcher.TNode(b.ID())
	if err != nil {
		return fmt.Errorf("%s:%s", head, err.Error())
	}
	rn, err := b.fetcher.RNode(b.ID())
	if err != nil {
		return fmt.Errorf("%s:%s", head, err.Error())
	}
	b.tNodes = tn
	b.rNodes = rn
	return nil
}
func (b *Base) setVariable(varbs variable.Set) error {
	if varbs != nil {
		b.variables = varbs
		return nil
	}
	return fmt.Errorf("setVariable: nil variables")
}

func (b *Base) ParseTrigger(T node.Item) error {
	distNode, err := b.findT(T, b.tNodes)
	if err != nil {
		return fmt.Errorf("parseTrigger:%s", err.Error())
	}
	logger.Debug("parseTrigger trigger node", "node", T.Value())
	logger.Debug("parseTrigger dist node", "node", distNode.Value())

	distNode.SetState(true)

	triggerVars, err := b.variableGen(T, distNode)
	if err != nil {
		return fmt.Errorf("parseTrigger:%s", err.Error())
	}

	if err := b.setVariable(triggerVars); err != nil {
		return fmt.Errorf("parseTrigger:%s", err.Error())
	}

	if err := b.mergeVariables(distNode); err != nil {
		return fmt.Errorf("parseTrigger:%s", err.Error())
	}

	return nil
}

func (b *Base) variableGen(T, dist node.Item) (variable.Set, error) {
	tMatches := b.varReg.FindAllStringSubmatch(T.Value(), -1)
	dMatches := b.varReg.FindAllStringSubmatch(dist.Value(), -1)
	if len(tMatches) != len(dMatches) {
		return nil, fmt.Errorf("variable count mismatch: %d != %d", len(tMatches), len(dMatches))
	}

	variables := make(variable.Set)
	for i, dMatch := range dMatches {
		if err := variableGen(i, dMatch, tMatches, T, variables); err != nil {
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
	hub := n.VariableSetHub()
	if len(hub) == 0 {
		// Fallback to current variables if hub is empty
		vars := n.Variables()
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
		for name, varItem := range n.Variables() {
			if _, exists := merged[name]; !exists {
				merged[name] = varItem.Copy()
			}
		}
	}
	return merged
}
func (b *Base) CheckAndExtract() error {
	var errs []error

	for _, t := range b.tNodes {

		if checker, ok := t.(node.Checkable); ok {
			var (
				keys      = t.VariableKeys()
				finalVars = make(variable.Set, len(keys))
			)
			for _, key := range keys {
				varb := b.variables[key]
				if varb == nil {
					errs = append(errs, fmt.Errorf("next: variable:%s not found", key))
					for k := range b.variables {
						logger.Error("next: variable", "name", k, "value", b.variables[k].Value())
					}
					continue
				}
				finalVars[key] = b.variables[key].Copy()
			}
			err := t.SetVariable(finalVars)
			if err != nil {
				errs = append(errs, fmt.Errorf("set variable failed: %w", err))
				for k := range b.variables {
					logger.Error("next: variable", "name", k, "value", b.variables[k].Value())
				}
				continue
			}
			if err = checker.Check(); err != nil {
				errs = append(errs, fmt.Errorf("check failed: %w", err))
			}
			continue
		}

		if extractor, ok := t.(node.Extractable); ok {
			var (
				keys      = t.VariableKeys()
				finalVars = make(variable.Set, len(keys))
			)
			// Get the extraction target variable - this should NOT be overridden
			extractTarget := extractor.ExtractTarget()
			for _, key := range keys {
				// Skip the extract target variable - it should remain unset for extraction
				if key == extractTarget {
					continue
				}
				varb := b.variables[key]
				if varb == nil {
					var msg string
					for k := range b.variables {
						msg += fmt.Sprintf("next: variable %s=%s\n", k, b.variables[k].Value())
					}
					for t := range b.tNodes {
						msg += fmt.Sprintf("next: trigger node %s\n", b.tNodes[t].Value())
					}
					errs = append(errs, fmt.Errorf("next: variable:%s not found:%s", key, msg))

					continue
				}
				finalVars[key] = b.variables[key].Copy()
			}
			err := t.SetVariable(finalVars)
			if err != nil {
				errs = append(errs, fmt.Errorf("set variable failed: %w", err))
				continue
			}
			varb, err := extractor.Extract()
			if err != nil {
				errs = append(errs, fmt.Errorf("extract failed: %w", err))
				continue
			}
			b.variables[varb.Name()] = varb
			if valueMap, ok := varb.(variable.ValueMap); ok {
				b.valueHash = valueMap.ValueMap()
				b.valueHashKey = varb.Name()
				logger.Debug("extractor returned stateHub", "size", len(b.valueHash))
			}
		}
	}

	return errors.Join(errs...)
}
func (b *Base) OK() (bool, error) {
	for _, t := range b.tNodes {
		if !t.OK() {
			return false, fmt.Errorf("trigger:%s is not activated", t.Value())
		}
		// Check if current variable values exist in this trigger's variableStateMap
		if _, ok := t.(node.Checkable); ok {
			continue
		}
		if _, ok := t.(node.Extractable); ok {
			continue
		}
		if ok, err := b.checkVariableStateMap(t); !ok {
			return false, err
		}
	}
	for _, varb := range b.variables {
		if !varb.OK() {
			return false, fmt.Errorf("variable:%s is empty", varb.Name())
		}
	}
	var rawSet = make(map[string]struct{})
	for _, t := range b.tNodes {
		var val, err = variable.ToRawText(t.Value(), b.variables, false)
		if err != nil {
			return false, fmt.Errorf("to raw text: %w", err)
		}
		if _, exists := rawSet[val]; exists {
			return false, fmt.Errorf("trigger:%s value:%s is not unique", t.Value(), val)
		}
		rawSet[val] = struct{}{}
	}
	return true, nil
}

// checkVariableStateMap verifies that the current variable values exist in the trigger's variableStateMap
func (b *Base) checkVariableStateMap(t node.Item) (bool, error) {
	stateMap := t.VariableStateMap()
	logger.Debug("checkVariableStateMap", "node", t.Value(), "variableStateMap", stateMap)
	var variableSet = make(variable.Set)
	for _, key := range t.VariableKeys() {
		if item, ok := b.variables[key]; ok {
			variableSet[key] = item
		} else {
			return false, fmt.Errorf("variable:%s not found", key)
		}
	}
	var state = node.GenVariableState(variableSet)
	if _, ok := t.VariableStateMap()[state]; !ok {
		return false, fmt.Errorf("variable state:%s not found", state)
	}
	return true, nil
}
func (b *Base) Next() error {
	var err []error
	for _, rn := range b.rNodes {
		var (
			keys      = rn.VariableKeys()
			finalVars = make(variable.Set, len(keys))
		)
		for _, key := range keys {
			varb := b.variables[key]
			if varb == nil {
				err = append(err, fmt.Errorf("next: variable:%s not found", key))
				continue
			}
			finalVars[key] = b.variables[key].Copy()
		}
		err = append(err, rn.SetVariable(finalVars))

		// If we have valueHash from extractor, create variable sets for each value
		if b.valueHash != nil {
			for value := range b.valueHash {
				finalVars[b.valueHashKey] = variable.New(b.valueHashKey, value)
				err = append(err, rn.SetVariable(finalVars))
			}
			logger.Debug("merged valueHash into result node", "node", rn.Value(), "states", len(b.valueHash))
		}
	}
	return errors.Join(err...)
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
func New(hashValue string, fetcher Fetcher, encoder node.Encoder, variableReg *regexp.Regexp) (Item, error) {
	head := "new container: "
	if err := utils.NotNull(
		"hashValue", hashValue,
		"fetcher", fetcher,
		"encoder", encoder,
		"variableReg", variableReg,
	); err != nil {
		return nil, fmt.Errorf("%s: %w", head, err)
	}
	return &Base{
		hashValue: hashValue,
		fetcher:   fetcher,
		encoder:   encoder,
		varReg:    variableReg,
	}, nil
}
