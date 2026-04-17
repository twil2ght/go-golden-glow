package node

import (
	"goldenglow/pkg/registry"
	"goldenglow/variable"
	"sync"
)

type Interface interface {
	Execute()
	Value() string
	ToTextWithNoVars(state string) string
}
type Node struct {
	value          string
	varSetRegistry registry.Registry[variable.Set]
	mu             sync.RWMutex
}

func (n *Node) Execute() {

}
func (n *Node) Value() string {
	return n.value
}
func (n *Node) ToTextWithNoVars(state string) string {
	varSet, _ := n.varSetRegistry.Get(state)
	e, _ := variable.ToRawText(n.value, varSet, false)
	return e
}
func New(value string) Interface {
	return &Node{
		value:          value,
		varSetRegistry: registry.New[variable.Set](),
		mu:             sync.RWMutex{},
	}
}
