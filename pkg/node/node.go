package node

import (
	"goldenglow/pkg/registry"
	"goldenglow/pkg/variable"
	"sync"
)

type Interface interface {
	Execute(state string)
	Value() string
	VarKeys() []string
	ToTextWithNoVars(state string) string
	VarSetRegistry() registry.Interface[variable.Set]
	NoState
}
type NoState interface {
	Activate()
	IsActivated() bool
}
type Node struct {
	value          string
	varSetRegistry registry.Interface[variable.Set]
	mu             sync.RWMutex
	activated      bool
}

func (n *Node) Activate() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.activated = true
}

func (n *Node) IsActivated() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.activated
}

func (n *Node) VarSetRegistry() registry.Interface[variable.Set] {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.varSetRegistry
}

func (n *Node) Execute(_ string) {

}
func (n *Node) Value() string {
	return n.value
}
func (n *Node) ToTextWithNoVars(state string) string {
	varSet, _ := n.varSetRegistry.Get(state)
	e, _ := variable.ToRawText(n.value, varSet, false)
	return e
}
func (n *Node) VarKeys() []string {
	return variable.VarReg.FindAllString(n.value, -1)
}
func New(value string) Interface {
	return &Node{
		value:          value,
		varSetRegistry: registry.New[variable.Set](),
		mu:             sync.RWMutex{},
	}
}
