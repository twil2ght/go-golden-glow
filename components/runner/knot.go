package runner

import (
	"errors"
	"fmt"
	"goldenglow/m"
	"goldenglow/node"
)

type Knot interface {
	Trigger() node.Item
	Trace() m.Hash
	TreeNode() *TreeNode
	SetTreeNode(node *TreeNode)
	State() string
	PrintTrace()
	Visit(node string)
}
type knot struct {
	trigger  node.Item
	trace    m.Hash
	visited  []string
	treeNode *TreeNode
	state    string
}

func (d *knot) Visit(node string) {
	d.visited = append(d.visited, node)
}
func (d *knot) PrintTrace() {
	for _, node := range d.visited {
		fmt.Println(node)
	}
}
func (d *knot) State() string {
	return d.state
}
func (d *knot) Trigger() node.Item {
	return d.trigger
}
func (d *knot) Trace() m.Hash {
	return d.trace
}
func (d *knot) TreeNode() *TreeNode {
	return d.treeNode
}
func (d *knot) SetTreeNode(node *TreeNode) {
	d.treeNode = node
}
func NewKnot(t, src node.Item, trace m.Hash) (Knot, error) {
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
		state:   node.GenVariableState(t.Variables()),
	}
	k.Visit(nodeValue)
	return k, nil
}
