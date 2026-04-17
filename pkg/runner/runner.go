package runner

import (
	"context"
	"goldenglow/components/queue"
	"goldenglow/container"
	"goldenglow/node"
	"goldenglow/node/template"
	"goldenglow/pkg/knot"
	"goldenglow/pkg/log"
	"time"
)

type Runner interface {
	Run(t node.Item, ctx context.Context)
}
type runner struct {
	workerNum        int
	ch               chan knot.Item
	containerFactory container.Factory
	queue            queue.Queue
}

var logger = log.Default()

func (r *runner) Run(t node.Item, ctx context.Context) {
	initKnot, _ := knot.New(t)
	r.ch <- initKnot

	for i := 0; i < r.workerNum; i++ {
		go r.worker(ctx)
	}

	<-ctx.Done()
}
func (r *runner) worker(ctx context.Context) {
	for {
		select {
		case e, ok := <-r.ch:
			if !ok {
				// Channel closed, stop worker
				return
			}
			if err := r.handler(e); err != nil {
				logger.Error("runner", "handler", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
func (r *runner) handler(k knot.Item) error {
	trigger := k.Trigger()
	_ = trigger.SetAndRegisterVars(trigger.GetVarSetByState(k.State()))
	rawValue, _ := trigger.ToTextWithoutVars()
	templateNodes, err := GetTemplates(trigger)
	if err != nil {
		return err
	}
	for _, tempN := range templateNodes {
		_ = tempN.SetAndRegisterVars(trigger.GetVarSetByState(rawValue))
		_ = tempN.Execute()
		cHashMap, err := r.containerFactory.Positioner().ContainerOf(tempN)
		if err != nil {
			return err
		}
		for hash := range cHashMap {
			c, err := r.containerFactory.New(hash)
			if err != nil {
				return err
			}
			ok, err := c.Do(tempN, k.State())
			if err != nil {
				return err
			}
			if ok {
				extraStates := c.ExtraStates()
				if len(extraStates) > 0 {
					for state := range extraStates {
						for _, rn := range c.RNode() {
							_ = rn.SetAndRegisterVars(rn.GetVarSetByState(state))
							NextKnot, err := knot.New(rn)
							if err != nil {
								return err
							}
							r.ch <- NextKnot
						}
					}
				} else {
					for _, rn := range c.RNode() {
						NextKnot, err := knot.New(rn)
						if err != nil {
							return err
						}
						r.ch <- NextKnot
					}
				}
			}
		}
	}
	return nil
}
func GetTemplates(t node.Item) (node.Set, error) {
	return template.DefaultCore().Get(t, false)
}
func New(containerFactory container.Factory, timeout time.Duration) Runner {
	return &runner{
		workerNum:        5,
		ch:               make(chan knot.Item, 1000),
		containerFactory: containerFactory,
	}
}
