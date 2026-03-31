package node

import (
	"fmt"
	"goldenglow/variable"
	"sort"
	"strings"
)

type Set map[string]Item

type AttrReader interface {
	Value() string
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
	VariableStateMap() map[string]map[string]bool
	VariableSetFromHub(state string) variable.Set
	MarkDone(state, cHash string)
	ToText() (string, error)
	Reset()
}
type Base struct {
	val            string
	state          bool
	variables      variable.Set
	parser         variable.Parser
	variableState  map[string]map[string]bool
	variableSetHub map[string]variable.Set
}

func (b *Base) ToText() (string, error) {
	e, err := b.parser(b.val, b.variables, false)
	if err != nil {
		return "", fmt.Errorf("%s parser error: %v", b.val, err)
	}
	return e, nil
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
func (b *Base) Variables() variable.Set {
	return variable.Copy(b.variables)
}
func (b *Base) VariableStateMap() map[string]map[string]bool {
	return b.variableState
}
func (b *Base) SetVariable(variables variable.Set) error {
	if variables == nil {
		return fmt.Errorf("SetVariable:nil")
	}
	b.variables = variables
	var variableState = GenVariableState(variables)
	if _, exists := b.variableState[variableState]; !exists {
		b.variableState[variableState] = map[string]bool{}
		b.variableSetHub[variableState] = variable.Copy(variables)
	}
	return nil
}
func (b *Base) VariableKeys() []string {
	return variable.VarReg.FindAllString(b.val, -1)
}
func (b *Base) VariableSetFromHub(state string) variable.Set {
	return b.variableSetHub[state]
}
func (b *Base) MarkDone(state, cHash string) {
	b.variableState[state][cHash] = true
}

func (b *Base) Reset() {
	b.variableState = make(map[string]map[string]bool)
	b.variableSetHub = make(map[string]variable.Set)
}

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
