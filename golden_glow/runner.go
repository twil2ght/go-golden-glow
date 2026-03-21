package goldenglow

//TODO docs
import (
	"fmt"
	"goldenglow/container"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/node/template"
	"goldenglow/pkg/log"
	"goldenglow/storage"
	"goldenglow/variable"
	"maps"
)

type Knot interface {
	Trigger() node.Item
	Trace() m.Hash
}
type knot struct {
	trigger node.Item
	trace   m.Hash
}

func (d *knot) Trigger() node.Item {
	return d.trigger
}
func (d *knot) Trace() m.Hash {
	return d.trace
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

type Engine interface {
	Run(node node.Item) error
}

type Base struct {
	Logger           log.Logger
	templateCore     template.Core
	containerFactory container.Factory
}

// SetLogger 替换日志器
func (b *Base) SetLogger(l log.Logger) error {
	if l != nil {
		b.Logger = l
		return nil
	}
	return fmt.Errorf("Base.SetLogger")
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

func New(logger log.Logger, db storage.Repository) Engine {
	// Default Config
	var (
		variableParser = variable.ToRawText
		nEncoder       = node.DefaultEncoder()

		nRegulator       node.Regulator
		nodeFactory      node.Factory
		fetcher          container.Fetcher
		positioner       container.Positioner
		templateCore     template.Core
		containerFactory container.Factory
		err              error
	)

	nRegulator, err = node.NewRegulator(storage.NewKVLite(db))
	if err != nil {
		logger.Error("failed to create node regulator", err)
		panic(err)
	}

	nodeFactory, err = node.NewFactory(variableParser, nRegulator)
	if err != nil {
		logger.Error("failed to create node factory", err)
		panic(err)
	}

	fetcher, err = container.NewFetcher(db, nodeFactory)
	if err != nil {
		logger.Error("failed to create fetcher", err)
		panic(err)
	}

	positioner, err = container.NewPositioner(db, nEncoder)
	if err != nil {
		logger.Error("failed to create positioner", err)
		panic(err)
	}

	templateCore, err = template.New(nil, variable.VarReg)
	if err != nil {
		logger.Error("failed to create template core", err)
		panic(err)
	}

	containerFactory, err = container.NewFactory(fetcher, nEncoder, positioner)
	if err != nil {
		logger.Error("failed to create container factory", err)
		panic(err)
	}

	return &Base{
		Logger:           logger,
		containerFactory: containerFactory,
		templateCore:     templateCore,
	}
}
func (b *Base) Run(input node.Item) error {

	initKnots, err := b.genKnots(input)
	if err != nil {
		return fmt.Errorf("run init step: %w", err)
	}

	taskCh := make(chan []Knot)
	resultCh := make(chan []Knot)
	done := make(chan error, 1)

	go func() {
		defer close(taskCh)
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
		defer close(resultCh)

		for knots := range taskCh {
			if len(knots) == 0 || knots == nil {
				done <- nil
				return
			}

			nextKnots, err := b.consume(knots)
			if err != nil {
				done <- err
				return
			}

			resultCh <- nextKnots
		}
	}()

	return <-done
}

func (b *Base) genKnots(n node.Item) ([]Knot, error) {
	nSet, _ := b.templateCore.Get(n)
	knots := make([]Knot, len(nSet))
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
		templateKnots, _ := b.genKnots(Item.Trigger())

		nextKnots = append(nextKnots, templateKnots...)
	}

	return nextKnots, nil
}
func (b *Base) consume(knots []Knot) ([]Knot, error) {
	nextKnots := make([]Knot, 0, len(knots))

	for _, Item := range knots {
		cHashMap, err := b.doTrigger(Item.Trigger())
		if err != nil {

		}
		for hashValue := range cHashMap {
			knots, err := b.doContainer(hashValue, Item)
			if err != nil {
			}
			nextKnots = append(nextKnots, knots...)
		}
	}

	return nextKnots, nil
}

// TODO  node.Execute
func (b *Base) doTrigger(n node.Item) (m.Hash, error) {
	n.SetState(true)
	err := n.Execute()
	if err != nil {

	}
	cHashMap, err := b.containerFactory.Positioner().ContainerOf(n)
	if err != nil {

	}
	return cHashMap, nil
}

func (b *Base) doContainer(hashValue string, T Knot) ([]Knot, error) {
	c, err := b.containerFactory.New(hashValue)
	if err != nil {
		return nil, err
	}
	err = c.Do(T.Trigger())
	if err != nil {
		return nil, err
	}
	triggers := c.RNode()
	var knots = make([]Knot, len(triggers))

	for _, t := range triggers {
		visited := make(map[string]struct{}, len(T.Trace()))
		maps.Copy(visited, T.Trace())
		k, err := NewKnot(t, m.Hash{})
		//TODO err
		if err != nil {
		}
		knots = append(knots, k)
	}
	return knots, nil
}
