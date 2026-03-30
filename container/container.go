package container

import (
	"errors"
	"fmt"
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
	hashValue string
	tNodes    node.Set
	rNodes    node.Set
	variables variable.Set
	fetcher   Fetcher
	encoder   node.Encoder
	varReg    *regexp.Regexp
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
		return false, err
	}

	err = b.ParseTrigger(T)
	if err != nil {
		return false, err
	}

	err = b.CheckAndExtract()
	if err != nil {
		return false, err
	}

	ok, err := b.OK()
	if !ok {
		if err != nil {
			return false, err
		}
		return false, nil
	}

	err = b.Next()
	if err != nil {
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
			err := t.SetVariable(b.variables)
			if err != nil {
				errs = append(errs, fmt.Errorf("set variable failed: %w", err))
				continue
			}
			if err = checker.Check(); err != nil {
				errs = append(errs, fmt.Errorf("check failed: %w", err))
			}
			continue
		}

		if extractor, ok := t.(node.Extractable); ok {
			err := t.SetVariable(b.variables)
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
	for key, varItem := range b.variables {
		if hash, ok := stateMap[key]; ok {
			if _, exists := hash[varItem.Value()]; !exists {
				logger.Debug("checkVariableStateMap variable value not exists", "variable", key, "value", varItem.Value())
				return false
			}
			logger.Debug("checkVariableStateMap variable value exists", "variable", key, "value", varItem.Value())
		}
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
