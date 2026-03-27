package node

import (
	"fmt"
	"goldenglow/variable"
	"strings"
)

type Set map[string]Item

type AttrReader interface {
	Value() string
	State() bool
	Variables() variable.Set
	VariableKeys() []string
}
type AttrWriter interface {
	SetState(state bool)
	SetVariable(variables variable.Set) error
}
type Executor interface {
	OK() bool
	Execute() error
}
type Item interface {
	Executor
	AttrReader
	AttrWriter
}
type Base struct {
	val       string
	state     bool
	variables variable.Set
	parser    variable.Parser
	regulator Regulator
}

func (b *Base) ToText() (string, error) {
	e, err := b.parser(b.val, b.variables, false)
	if err != nil {
		return "", fmt.Errorf("%s parser error: %v", b.val, err)
	}
	return b.regulator.Do(e), nil
}
func (b *Base) SetState(state bool) {
	b.state = state
}
func (b *Base) OK() bool {
	return b.state
}
func (b *Base) Execute() error {
	if strings.HasPrefix(b.val, "[node") {
		return fmt.Errorf("node: %s not implemented", b.val)
	}
	return nil
}
func (b *Base) Value() string {
	return b.val
}
func (b *Base) State() bool {
	return b.state
}
func (b *Base) Variables() variable.Set {
	return variable.Copy(b.variables)
}

func (b *Base) SetVariable(variables variable.Set) error {
	if variables == nil {
		return fmt.Errorf("SetVariable:nil")
	}
	b.variables = variables
	return nil
}
func (b *Base) VariableKeys() []string {
	return variable.VarReg.FindAllString(b.val, -1)
}
