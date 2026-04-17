package runner

import (
	"goldenglow/container"
	"goldenglow/node"
	"goldenglow/node/template"
	"goldenglow/pkg/knot"
	"goldenglow/pkg/log"
	"sync"
	"time"
)

type Runner interface {
	Run(t node.Item)
}
type runner struct {
	workerNum        int
	ch               chan knot.Item
	stopCh           chan struct{}
	containerFactory container.Factory
	timeout          time.Duration
	stopOnce         *sync.Once
}

var logger = log.Default()

func (r *runner) Run(t node.Item) {
	initKnot, _ := knot.New(t)
	r.ch <- initKnot

	for i := 0; i < r.workerNum; i++ {
		go r.worker(i)
	}
	<-r.stopCh
}
func (r *runner) worker(_ int) {
	// Process until channel is empty and timeout has passed
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
		case <-time.After(r.timeout):
			// Timeout: channel has been empty for the specified duration
			// Use sync.Once to ensure thread-safe single execution
			r.stopOnce.Do(func() {
				close(r.ch)
				r.stopCh <- struct{}{}
			})
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
		stopOnce:         &sync.Once{},
		stopCh:           make(chan struct{}),
		containerFactory: containerFactory,
		timeout:          timeout,
	}
}
