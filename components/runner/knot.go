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
}
type knot struct {
	trigger  node.Item
	trace    m.Hash
	treeNode *TreeNode
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
func NewKnot(t, src node.Item, trace m.Hash, encoder node.Encoder) (Knot, error) {
	if t == nil {
		return nil, fmt.Errorf("NewKnot: trigger==nil")
	}
	if trace == nil {
		trace = m.Hash{}
	}
	var (
		nodeValue = t.Value()
		encoded   = encoder.Do(nodeValue)
	)
	if _, ok := trace[encoded]; ok && src == nil {
		return nil, errors.New("duplicate node")
	}
	trace[encoded] = struct{}{}
	return &knot{
		trigger: t,
		trace:   trace,
	}, nil
}
