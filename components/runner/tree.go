package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// TreeNode represents a node in the execution tree
type TreeNode struct {
	Name     string      `json:"name"`
	Depth    int         `json:"depth"`
	IsLast   bool        `json:"isLast"`
	Parent   *TreeNode   `json:"-"`
	Children []*TreeNode `json:"children"`
}

// TreeBuilder collects all nodes during execution and prints at the end
type TreeBuilder struct {
	root  *TreeNode
	nodes []*TreeNode
}

func NewTreeBuilder() *TreeBuilder {
	return &TreeBuilder{
		nodes: make([]*TreeNode, 0),
	}
}

func (tb *TreeBuilder) AddNode(name string, depth int, parent *TreeNode) *TreeNode {
	n := &TreeNode{
		Name:   name,
		Depth:  depth,
		Parent: parent,
	}
	if parent != nil {
		parent.Children = append(parent.Children, n)
	}
	tb.nodes = append(tb.nodes, n)
	return n
}

func (tb *TreeBuilder) Print() {
	fmt.Println("═══ EXECUTION TREE ═══")
	for _, treeNode := range tb.nodes {
		if treeNode.Parent == nil {
			// Root level nodes
			tb.printNode(treeNode)
		}
	}
	fmt.Println("═══ EXECUTION END ═══")
}

// ToJSON exports the tree structure as JSON for HTML visualization
func (tb *TreeBuilder) ToJSON() ([]byte, error) {
	rootNodes := make([]*TreeNode, 0)
	for _, treeNode := range tb.nodes {
		if treeNode.Parent == nil {
			rootNodes = append(rootNodes, treeNode)
		}
	}
	if len(rootNodes) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(rootNodes)
}

// SaveToFile saves the tree JSON to a file for HTML visualization
func (tb *TreeBuilder) SaveToFile(filepath string) error {
	jsonData, err := tb.ToJSON()
	if err != nil {
		return fmt.Errorf("SaveToFile: marshal: %w", err)
	}
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return fmt.Errorf("SaveToFile: write: %w", err)
	}
	return nil
}

func (tb *TreeBuilder) printNode(node *TreeNode) {
	prefix := tb.buildPrefix(node.Depth, node.IsLast)
	fmt.Printf("%s%s\n", prefix, node.Name)
	for i, child := range node.Children {
		child.IsLast = i == len(node.Children)-1
		tb.printNode(child)
	}
}

func (tb *TreeBuilder) buildPrefix(depth int, isLast bool) string {
	if depth == 0 {
		if isLast {
			return "└── "
		}
		return "├── "
	}
	var sb strings.Builder
	for i := 0; i < depth; i++ {
		sb.WriteString("│   ")
	}
	if isLast {
		sb.WriteString("└── ")
	} else {
		sb.WriteString("├── ")
	}
	return sb.String()
}
