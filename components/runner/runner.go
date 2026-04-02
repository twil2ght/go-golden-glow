package runner

//TODO docs
import (
	"errors"
	"fmt"
	"goldenglow/container"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/node/template"
	"goldenglow/pkg/log"
	"maps"
	"strings"
)

var logger = log.Default()

// TreeNode represents a node in the execution tree
type TreeNode struct {
	Name     string
	Depth    int
	IsLast   bool
	Parent   *TreeNode
	Children []*TreeNode
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
func NewKnot(t node.Item, trace m.Hash) (Knot, error) {
	if t == nil {
		return nil, fmt.Errorf("NewKnot: trigger==nil")
	}
	if trace == nil {
		trace = m.Hash{}
	}
	return &knot{
		trigger: t,
		trace:   trace,
	}, nil
}

type Instance interface {
	Run(node node.Item) error
}

type Base struct {
	templateCore     template.Core
	containerFactory container.Factory
	treeBuilder      *TreeBuilder
}

func (b *Base) SetContainerFactory(f container.Factory) error {
	if f != nil {
		b.containerFactory = f
		return nil
	}
	return fmt.Errorf("Base.SetContainerFactory")
}
func (b *Base) SetTemplateCore(f template.Core) error {
	if f != nil {
		b.templateCore = f
		return nil
	}
	return fmt.Errorf("Base.SetTemplateCore")
}

func New(cf container.Factory, tc template.Core) Instance {
	return &Base{
		containerFactory: cf,
		templateCore:     tc,
	}
}

func (b *Base) Run(input node.Item) error {
	b.treeBuilder = NewTreeBuilder()
	b.treeBuilder.AddNode("📥 INPUT: "+input.Value(), 0, nil)

	initKnots, err := b.genKnots(input)
	if err != nil {
		return fmt.Errorf("run init step: %w", err)
	}

	// 带缓冲通道，避免无缓冲导致的立即阻塞
	taskCh := make(chan []Knot, 1000)
	resultCh := make(chan []Knot, 1000)
	done := make(chan error, 1)

	go func() {
		defer func() {
			close(taskCh)
		}()

		// 初始化数据推入通道
		resultCh <- initKnots

		for knots := range resultCh {
			nextKnots, err := b.produce(knots)
			if err != nil {
				done <- err
				return
			}
			if len(nextKnots) == 0 {
				taskCh <- nil
				return
			}
			taskCh <- nextKnots
		}
	}()

	go func() {
		defer func() {
			close(resultCh)
		}()

		for knots := range taskCh {
			if knots == nil || len(knots) == 0 {
				done <- nil
				return
			}
			nextKnots, err := b.consume(knots)
			if err != nil {
				logger.Debug("consume: some knots failed", "error", err)
			}
			resultCh <- nextKnots
		}
	}()

	err = <-done
	if err != nil {
		b.containerFactory.ResetNodePool()
		return err
	}

	// Print the tree after execution is complete
	b.treeBuilder.Print()

	// Reset all nodes' variableState and variableSetHub after each run
	b.containerFactory.ResetNodePool()

	return nil
}

func (b *Base) genKnots(n node.Item) ([]Knot, error) {
	nSet, err := b.templateCore.Get(n)
	if err != nil {
		return nil, err
	}
	logger.Debug("get templates", "templates_amount", len(nSet), "from", n.Value())
	knots := make([]Knot, 0, len(nSet))
	for _, n := range nSet {
		k, err := NewKnot(n, m.Hash{})
		if err != nil {
			return nil, err
		}
		knots = append(knots, k)
	}
	return knots, nil
}

func (b *Base) produce(knots []Knot) ([]Knot, error) {
	nextKnots := make([]Knot, 0, len(knots))
	for _, Item := range knots {
		var (
			encoder   = b.containerFactory.Encoder()
			nodeValue = Item.Trigger().Value()
			encoded   = encoder.Do(nodeValue)
		)
		if _, ok := Item.Trace()[encoded]; ok {
			continue
		}
		Item.Trace()[encoded] = struct{}{}

		// Get parent tree node - if nil, this is an initial knot from input
		parentTreeNode := Item.TreeNode()
		if parentTreeNode == nil {
			// Create tree node for initial knots (direct children of INPUT)
			parentTreeNode = b.treeBuilder.AddNode("🌿 "+Item.Trigger().Value(), 1, nil)
			Item.SetTreeNode(parentTreeNode)
		}

		templateKnots, err := b.genKnots(Item.Trigger())
		if err != nil {
			return nil, err
		}
		parentValue := Item.Trigger().Value()
		for _, tk := range templateKnots {
			if tk.Trigger().Value() == parentValue {
				// This is the parent node itself - reuse existing tree node
				tk.SetTreeNode(parentTreeNode)
			} else {
				// This is a new template-generated node - create tree node
				tk.SetTreeNode(b.treeBuilder.AddNode("🌿 "+tk.Trigger().Value(), parentTreeNode.Depth+1, parentTreeNode))
			}
			nextKnots = append(nextKnots, tk)
		}
	}
	return nextKnots, nil
}

func (b *Base) consume(knots []Knot) ([]Knot, error) {
	nextKnots := make([]Knot, 0, len(knots))
	var errs []error
	for _, Item := range knots {
		triggerNode := Item.Trigger()
		triggerValue := triggerNode.Value()

		// Get or create tree node for this trigger
		var treeNode *TreeNode
		if Item.TreeNode() != nil {
			treeNode = Item.TreeNode()
		} else {
			treeNode = b.treeBuilder.AddNode("🌿 "+triggerValue, 0, nil)
			Item.SetTreeNode(treeNode)
		}

		cHashMap, err := b.processTrigger(triggerNode)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		for hashValue := range cHashMap {
			for state, doneMap := range triggerNode.VariableStateMap() {
				var done = doneMap[hashValue]
				if done {
					continue
				}
				err := triggerNode.SetVariable(triggerNode.VariableSetFromHub(state))
				triggerNode.MarkDone(state, hashValue)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				newKnots, err := b.processContainer(hashValue, Item, treeNode)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				nextKnots = append(nextKnots, newKnots...)
			}
		}
	}
	return nextKnots, errors.Join(errs...)
}

func (b *Base) processTrigger(n node.Item) (m.Hash, error) {
	var errGroup []error
	n.SetState(true)
	for state := range n.VariableStateMap() {
		if n.VariableStateExecute()[state] {
			continue
		}
		_ = n.SetVariable(n.VariableSetFromHub(state))
		n.MarkExecuteState(state)
		if err := n.Execute(); err != nil {
			errGroup = append(errGroup, err)
		}
	}
	cHashMap, err := b.containerFactory.Positioner().ContainerOf(n)
	if err != nil {
		errGroup = append(errGroup, err)
	}
	return cHashMap, errors.Join(errGroup...)
}

func (b *Base) processContainer(hashValue string, T Knot, parentTreeNode *TreeNode) ([]Knot, error) {
	c, err := b.containerFactory.New(hashValue)
	if err != nil {
		return nil, err
	}
	ok, err := c.Do(T.Trigger())
	if err != nil {
		return nil, err
	}
	results := c.RNode()
	triggers := c.TNode()

	// Add container node to tree
	containerNode := b.treeBuilder.AddNode("📦 CONTAINER [ID: "+hashValue+"]", parentTreeNode.Depth+1, parentTreeNode)
	for _, t := range triggers {
		raw, err := t.ToText()
		if err != nil {
			raw = "?"
		}
		_ = b.treeBuilder.AddNode("🔗 TRIGGER: "+t.Value()+fmt.Sprintf("(%s)", raw), containerNode.Depth+1, containerNode)
	}
	if !ok {
		return nil, nil
	}

	var knots = make([]Knot, 0, len(results))
	for _, t := range results {
		resultValue := t.Value()
		raw, err := t.ToText()
		if err != nil {
			raw = "?"
		}
		// Add result node to tree
		resultNode := b.treeBuilder.AddNode("🎯 RESULT: "+resultValue+fmt.Sprintf("(%s)", raw), containerNode.Depth+1, containerNode)

		visited := make(map[string]struct{}, len(T.Trace()))
		maps.Copy(visited, T.Trace())
		k, err := NewKnot(t, m.Hash{})
		if err != nil {
			continue
		}
		// Link the knot to the result tree node
		k.SetTreeNode(resultNode)
		knots = append(knots, k)
	}
	return knots, nil
}
