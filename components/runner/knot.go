package runner

import (
	"errors"
	"fmt"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/variable"
	"maps"
)

type Knot interface {
	Trigger() node.Item
	Trace() m.Hash
	TreeNode() *TreeNode
	SetTreeNode(node *TreeNode)
	State() string
	SetState(str string)
}
type knot struct {
	trigger  node.Item
	trace    m.Hash
	treeNode *TreeNode
	state    string
}

func (d *knot) State() string {
	return d.state
}
func (d *knot) SetState(str string) {
	d.state = str
}
func (d *knot) Trigger() node.Item {
	return d.trigger
}
func (d *knot) Trace() m.Hash {
	visited := make(map[string]struct{}, len(d.trace))
	maps.Copy(visited, d.trace)
	return visited
}
func (d *knot) TreeNode() *TreeNode {
	return d.treeNode
}
func (d *knot) SetTreeNode(node *TreeNode) {
	d.treeNode = node
}
func NewKnot(t, src node.Item, trace m.Hash, set variable.Set) (Knot, error) {
	if t == nil {
		return nil, fmt.Errorf("NewKnot: trigger==nil")
	}
	if trace == nil {
		trace = m.Hash{}
	}
	var (
		nodeValue = t.Value()
	)
	if _, ok := trace[nodeValue]; ok && src == nil {
		return nil, errors.New("duplicate node" + nodeValue + fmt.Sprintf("(%+v)", trace))
	}
	trace[nodeValue] = struct{}{}

	k := &knot{
		trigger: t,
		trace:   trace,
		state:   node.GenVariableState(set),
	}
	return k, nil
}
