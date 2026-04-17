package runner

import (
	"context"
	"goldenglow/container"
	"goldenglow/node"
	"goldenglow/node/template"
	"goldenglow/pkg/knot"
	"goldenglow/pkg/log"
	"time"
)

type Queue interface {
	Get() node.Item
}
type Runner interface {
	Run(t node.Item, ctx context.Context)
}
type runner struct {
	workerNum        int
	ch               chan knot.Item
	containerFactory container.Factory
	queue            Queue
}

var logger = log.Default()

func (r *runner) Run(t node.Item, ctx context.Context) {
	for i := 0; i < r.workerNum; i++ {
		go r.worker(ctx)
	}

	go r.watchIdle(ctx, 1*time.Second, 500*time.Millisecond)

	<-ctx.Done()
}
func (r *runner) watchIdle(ctx context.Context, checkInterval time.Duration, timeout time.Duration) {
	ticker := time.NewTicker(checkInterval) // Check every second
	defer ticker.Stop()

	var idleStartTime time.Time
	isIdle := false

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check if channel is empty
			if len(r.ch) == 0 {
				if !isIdle {
					// Just became idle
					isIdle = true
					idleStartTime = time.Now()
				}

				// Check if idle time has exceeded threshold
				if time.Since(idleStartTime) > timeout { // 5 seconds idle threshold
					r.onIdle()
					isIdle = false // Reset to trigger again later if needed
				}
			} else {
				// Channel is not empty, reset idle state
				isIdle = false
			}
		}
	}
}

// onIdle is called when the channel has been empty for the specified duration
func (r *runner) onIdle() {
	initKnot, _ := knot.New(r.queue.Get())
	r.ch <- initKnot
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
