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
	"goldenglow/utils"
	"maps"
)

var logger = log.Default()

type Instance interface {
	Run(node node.Item) error
}

type Base struct {
	templateCore     template.Core
	containerFactory container.Factory
	treeBuilder      *TreeBuilder
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

	initKnots, err := b.genKnots(input, m.Hash{})
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
		// 初始化数据推入通道
		taskCh <- initKnots
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

	// Save tree to JSON file for HTML visualization
	treeFile := utils.RootDir + "/tree_output/tree.json"
	if err := b.treeBuilder.SaveToFile(treeFile); err != nil {
		logger.Debug("Failed to save tree file", "error", err)
	} else {
		logger.Info("Tree saved", "file", treeFile)
	}

	// Reset all nodes' variableState and variableSetHub after each run
	b.containerFactory.ResetNodePool()

	return nil
}

func (b *Base) genKnots(src node.Item, trace m.Hash) ([]Knot, error) {
	nSet, err := b.templateCore.Get(src)
	if err != nil {
		return nil, err
	}
	logger.Debug("get templates", "templates_amount", len(nSet), "from", src.Value())
	knots := make([]Knot, 0, len(nSet))
	for _, n := range nSet {
		k, err := NewKnot(n, src, trace, b.containerFactory.Encoder())
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
		//var (
		//	encoder   = b.containerFactory.Encoder()
		//	nodeValue = Item.Trigger().Value()
		//	encoded   = encoder.Do(nodeValue)
		//)
		//if _, ok := Item.Trace()[encoded]; ok {
		//	b.treeBuilder.AddNode("🌿 "+"duplicated and skip", Item.TreeNode().Depth+1, Item.TreeNode())
		//	continue
		//}
		//Item.Trace()[encoded] = struct{}{}

		// Get parent tree node - if nil, this is an initial knot from input
		parentTreeNode := Item.TreeNode()
		if parentTreeNode == nil {
			// Create tree node for initial knots (direct children of INPUT)
			parentTreeNode = b.treeBuilder.AddNode("🌿 "+Item.Trigger().Value(), 1, nil)
			Item.SetTreeNode(parentTreeNode)
		}

		templateKnots, err := b.genKnots(Item.Trigger(), Item.Trace())
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
	results := c.RNode()
	triggers := c.TNode()

	// Add container node to tree
	containerNode := b.treeBuilder.AddNode("📦 CONTAINER [ID: "+hashValue+"]", parentTreeNode.Depth+1, parentTreeNode)
	for _, t := range triggers {
		raw, err := t.ToText()
		if err != nil {
			raw = "?"
		}
		if t.Value() == T.Trigger().Value() {
			_ = b.treeBuilder.AddNode("🚀 TRIGGER: "+t.Value()+fmt.Sprintf("(%s)", raw), containerNode.Depth+1, containerNode)
			continue
		}
		_ = b.treeBuilder.AddNode("🔗 TRIGGER: "+t.Value()+fmt.Sprintf("(%s)", raw), containerNode.Depth+1, containerNode)
	}
	if !ok {
		if err != nil {
			b.treeBuilder.AddNode("❌️ REASON: "+err.Error(), parentTreeNode.Depth+1, parentTreeNode)
		}
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
		k, err := NewKnot(t, nil, visited, b.containerFactory.Encoder())
		if err != nil {
			continue
		}
		// Link the knot to the result tree node
		k.SetTreeNode(resultNode)
		knots = append(knots, k)
	}
	return knots, nil
}
