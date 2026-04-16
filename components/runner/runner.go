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

	initKnots, err := b.genKnots(input, m.Hash{}, true)
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

func (b *Base) genKnots(src node.Item, trace m.Hash, specific bool) ([]Knot, error) {
	nSet, err := b.templateCore.Get(src, specific)
	if err != nil {
		return nil, err
	}
	knots := make([]Knot, 0, len(nSet))
	var raw, _ = src.ToText()
	for _, n := range nSet {
		var s node.Item
		if n.Value() == src.Value() {
			s = src
		} else {
			s = nil
		}
		visited := make(map[string]struct{}, len(trace))
		maps.Copy(visited, trace)
		k, err := NewKnot(n, s, visited, n.VariableSetFromHub(raw))
		if err != nil {
			continue
		}
		knots = append(knots, k)
	}
	return knots, nil
}
func (b *Base) TempKnot(src node.Item, trace m.Hash) (Knot, error) {
	nSet, err := b.templateCore.Get(src, false)
	if err != nil {
		return nil, err
	}
	nMap := b.templateCore.RemoveTar(src, nSet)
	if len(nMap) == 0 {
		return nil, errors.New("no knots found")
	}
	var n node.Item
	for _, e := range nMap {
		n = e
		break
	}
	var raw, _ = src.ToText()
	var s node.Item
	if n.Value() == src.Value() {
		s = src
	} else {
		s = nil
	}
	visited := make(map[string]struct{}, len(trace))
	maps.Copy(visited, trace)
	k, err := NewKnot(n, s, visited, n.VariableSetFromHub(raw))
	if err != nil {
		return nil, err
	}
	return k, nil
}
func (b *Base) produce(knots []Knot) ([]Knot, error) {
	nextKnots := make([]Knot, 0, len(knots))
	for _, Item := range knots {
		// Get parent tree node - if nil, this is an initial knot from input
		parentTreeNode := Item.TreeNode()
		if parentTreeNode == nil {
			// Create tree node for initial knots (direct children of INPUT)
			parentTreeNode = b.treeBuilder.AddNode("🌿 "+Item.Trigger().Value(), 1, nil)
			Item.SetTreeNode(parentTreeNode)
		}
		templateKnots, err := b.genKnots(Item.Trigger(), Item.Trace(), false)
		Item.Trigger().SetVariable(Item.Trigger().VariableSetFromHub(Item.State()))
		var raw, _ = Item.Trigger().ToText()
		b.treeBuilder.AddNode("⭐ "+Item.Trigger().Value()+"("+raw+")", parentTreeNode.Depth+1, parentTreeNode)
		b.treeBuilder.AddNode("⭐ "+Item.Trigger().Value()+"("+Item.State()+")", parentTreeNode.Depth+1, parentTreeNode)
		if err != nil {
			return nil, err
		}
		if len(templateKnots) == 0 {
			return nil, fmt.Errorf("no knots generated for %s", Item.Trigger().Value())
		}

		parentValue := Item.Trigger().Value()
		if len(templateKnots) <= 1 {
			ch, _ := b.processTrigger(templateKnots[0].Trigger())
			if len(ch) == 0 {
				tempKnot, err := b.TempKnot(templateKnots[0].Trigger(), Item.Trace())
				if err == nil {
					templateKnots = []Knot{tempKnot}
				}
			}
		}
		for _, tk := range templateKnots {
			tk.Trigger().SetVariable(tk.Trigger().VariableSetFromHub(raw))
			tk.SetState(node.GenVariableState(tk.Trigger().Variables()))
			if tk.Trigger().Value() == parentValue {
				// This is the parent node itself - reuse existing tree node
				tk.SetTreeNode(parentTreeNode)
			} else {
				var raw, _ = tk.Trigger().ToText()
				// This is a new template-generated node - create tree node
				tk.SetTreeNode(b.treeBuilder.AddNode("🌿 "+tk.Trigger().Value()+"("+raw+")", parentTreeNode.Depth+1, parentTreeNode))
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
			state := Item.State()
			err := triggerNode.SetVariable(triggerNode.VariableSetFromHub(state))
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
	return nextKnots, errors.Join(errs...)
}

func (b *Base) processTrigger(n node.Item) (m.Hash, error) {
	var errGroup []error
	n.SetState(true)
	var prevVariables = n.Variables()
	defer func() {
		_ = n.SetVariable(prevVariables)
	}()
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
	ok, err := c.Do(T.Trigger(), T.State())
	results := c.RNode()
	triggers := c.TNode()

	// Add container node to tree
	b.treeBuilder.AddNode("⚙️ State ["+T.State()+"]", parentTreeNode.Depth+1, parentTreeNode)
	containerNode := b.treeBuilder.AddNode("📦 CONTAINER [ID: "+hashValue+"]", parentTreeNode.Depth+1, parentTreeNode)
	for _, t := range triggers {
		raw, _ := t.ToText()
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
		variableSet := t.Variables()
		raw, _ := t.ToText()
		// Add result node to tree
		resultNode := b.treeBuilder.AddNode("🎯 RESULT: "+resultValue+fmt.Sprintf("(%s)", raw), containerNode.Depth+1, containerNode)

		k, err := NewKnot(t, nil, T.Trace(), variableSet)
		if err != nil {
			b.treeBuilder.AddNode("⚠️ RESULT ERROR: "+err.Error()+fmt.Sprintf("(%s)", raw), containerNode.Depth+1, containerNode)
			continue
		}
		// Link the knot to the result tree node
		k.SetTreeNode(resultNode)
		knots = append(knots, k)
	}
	return knots, nil
}
