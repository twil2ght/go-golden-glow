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
)

var logger = log.Default()

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

type Instance interface {
	Run(node node.Item) error
}

type Base struct {
	templateCore     template.Core
	containerFactory container.Factory
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
	initKnots, err := b.genKnots(input)
	if err != nil {
		return fmt.Errorf("run init step: %w", err)
	}
	logger.Debug("start to run", "initial_knots_amount", len(initKnots))
	// 带缓冲通道，避免无缓冲导致的立即阻塞
	taskCh := make(chan []Knot, 1)
	resultCh := make(chan []Knot, 1)
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
	var errs []error

	for _, Item := range knots {
		cHashMap, err := b.processTrigger(Item.Trigger())
		if err != nil {
			errs = append(errs, err)
			continue
		}
		for hashValue := range cHashMap {
			knots, err := b.processContainer(hashValue, Item)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			nextKnots = append(nextKnots, knots...)
		}
	}

	return nextKnots, errors.Join(errs...)
}

func (b *Base) processTrigger(n node.Item) (m.Hash, error) {
	var errGroup []error
	n.SetState(true)
	err := n.Execute()
	if err != nil {
		errGroup = append(errGroup, err)
	}
	cHashMap, err := b.containerFactory.Positioner().ContainerOf(n)
	if err != nil {
		errGroup = append(errGroup, err)
	}
	return cHashMap, errors.Join(errGroup...)
}

func (b *Base) processContainer(hashValue string, T Knot) ([]Knot, error) {
	c, err := b.containerFactory.New(hashValue)
	if err != nil {
		return nil, err
	}
	err = c.Do(T.Trigger())
	if err != nil {
		return nil, err
	}
	triggers := c.RNode()
	var knots = make([]Knot, 0, len(triggers))

	for _, t := range triggers {
		visited := make(map[string]struct{}, len(T.Trace()))
		maps.Copy(visited, T.Trace())
		k, err := NewKnot(t, m.Hash{})
		if err != nil {
			continue
		}
		knots = append(knots, k)
	}
	return knots, nil
}
