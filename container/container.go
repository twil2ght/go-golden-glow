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
		return false, nil
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
	for _, t := range b.tNodes {
		if t.Value() == excludeNode.Value() {
			continue
		}
		for name, varItem := range t.Variables() {
			if _, exists := b.variables[name]; !exists {
				b.variables[name] = varItem
			}
		}
	}
	return nil
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
			return false, nil
		}
		// Check if current variable values exist in this trigger's variableStateMap
		if _, ok := t.(node.Checkable); ok {
			continue
		}
		if _, ok := t.(node.Extractable); ok {
			continue
		}
		if !b.checkVariableStateMap(t) {
			return false, nil
		}
	}
	for _, varb := range b.variables {
		if !varb.OK() {
			return false, fmt.Errorf("variable:%s is empty", varb.Name())
		}
	}
	return true, nil
}

// checkVariableStateMap verifies that the current variable values exist in the trigger's variableStateMap
func (b *Base) checkVariableStateMap(t node.Item) bool {
	stateMap := t.VariableStateMap()
	logger.Debug("checkVariableStateMap", "node", t.Value(), "variableStateMap", stateMap)
	var variableSet = make(variable.Set)
	for _, key := range t.VariableKeys() {
		if item, ok := b.variables[key]; ok {
			variableSet[key] = item
		} else {
			return false
		}
	}
	var state = node.GenVariableState(variableSet)
	if _, ok := t.VariableStateMap()[state]; !ok {
		return false
	}
	return true
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
