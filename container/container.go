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
	"sync"
)

var logger = log.Default()

type Item interface {
	ID() string
	TNode() node.Set
	RNode() node.Set
	Variables() variable.Set
	Do(external node.Item, state string) (bool, error)
	ExtraStates() m.Hash
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
	mu           *sync.RWMutex
}

func (b *Base) ExtraStates() m.Hash {
	return b.valueHash
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
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.variables
}
func (b *Base) SetVariable(variables variable.Set) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if variables != nil {
		b.variables = variables
		return nil
	}
	return fmt.Errorf("setVariable: nil variables")
}
func (b *Base) Do(T node.Item, state string) (bool, error) {
	err := b.SetNode()
	_ = T.SetAndRegisterVars(T.GetVarSetByState(state))
	GivenSet := T.Vars()
	if err != nil {
		logger.Debug("setNode failed", "id", b.ID(), "error", err)
		return false, err
	}

	err = b.ParseTrigger(T, GivenSet)
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

	err = b.PassDownVariablesToResults()
	if err != nil {
		logger.Debug("next failed", "id", b.ID(), "error", err)
		return false, err
	}
	logger.Debug("container is ok", "id", b.ID())
	return true, nil
}
func (b *Base) SetNode() error {
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
func (b *Base) ParseTrigger(T node.Item, GivenSet variable.Set) error {
	distNode, err := b.findT(T, b.tNodes)
	if err != nil {
		return fmt.Errorf("parseTrigger:%s", err.Error())
	}
	logger.Debug("parseTrigger trigger node", "node", T.Value())
	logger.Debug("parseTrigger dist node", "node", distNode.Value())

	distNode.SetState(true)

	triggerVars, err := b.variableGen(T, distNode, GivenSet)
	if err != nil {
		return fmt.Errorf("parseTrigger:%s", err.Error())
	}

	if err := b.SetVariable(triggerVars); err != nil {
		return fmt.Errorf("parseTrigger:%s", err.Error())
	}

	if err := b.mergeVariables(distNode); err != nil {
		return fmt.Errorf("parseTrigger:%s", err.Error())
	}

	return nil
}

func (b *Base) CheckAndExtract() error {
	var errs []error

	for _, t := range b.tNodes {
		if err := b.processNode(t); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
func (b *Base) OK() (bool, error) {
	for _, t := range b.tNodes {
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
	for _, variableItem := range b.variables {
		if !variableItem.OK() {
			return false, fmt.Errorf("variable:%s is empty", variableItem.Name())
		}
	}
	return true, nil
}

func (b *Base) PassDownVariablesToResults() error {
	var err []error
	for _, rn := range b.rNodes {
		var (
			keys      = rn.VarKeys()
			finalVars = make(variable.Set, len(keys))
		)
		for _, key := range keys {
			variableItem := b.variables[key]
			if variableItem == nil {
				err = append(err, fmt.Errorf("next: variable:%s not found", key))
				continue
			}
			finalVars[key] = b.variables[key].Copy()
		}
		err = append(err, rn.SetAndRegisterVars(finalVars))
	}
	return errors.Join(err...)
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
		mu:        &sync.RWMutex{},
	}, nil
}
