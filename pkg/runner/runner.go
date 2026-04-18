package runner

import (
	"context"
	"goldenglow/m"
	"goldenglow/pkg/container"
	"goldenglow/pkg/container/positioner"
	"goldenglow/pkg/knot"
	"goldenglow/pkg/log"
	"goldenglow/pkg/node"
	"goldenglow/pkg/node/template"
	"sync"
	"time"
)

type Queue[T any] interface {
	Get() (item T, shutdown bool)
	Add(item T)
	Len() int
	Shutdown()
}
type Runner interface {
	Run(ctx context.Context)
}
type runner struct {
	workerNum     int
	wg            *sync.WaitGroup
	stopChan      chan struct{}
	externalQueue Queue[node.Interface]
	knotQueue     Queue[knot.Interface]
}

var logger = log.Default()

func (r *runner) Run(ctx context.Context) {
	r.wg.Add(r.workerNum + 1)

	for i := 0; i < r.workerNum; i++ {
		go r.worker(i)
	}

	go r.watchIdle(ctx, 1*time.Second, 500*time.Millisecond)

	r.wg.Wait()
}
func (r *runner) watchIdle(ctx context.Context, checkInterval time.Duration, timeout time.Duration) {
	defer r.wg.Done()
	ticker := time.NewTicker(checkInterval) // Check every second
	defer ticker.Stop()

	var idleStartTime time.Time
	isIdle := false

	for {
		select {
		case <-ctx.Done():
			r.knotQueue.Shutdown()
			return
		case <-r.stopChan:
			r.knotQueue.Shutdown()
			return
		case <-ticker.C:
			// Check if channel is empty
			if r.knotQueue.Len() == 0 {
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
	n, shutdown := r.externalQueue.Get()
	if !shutdown {
		initKnot := knot.New(n, "")
		if initKnot != nil {
			r.knotQueue.Add(initKnot)
		}
	} else {
		close(r.stopChan)
	}
}
func (r *runner) worker(_ int) {
	defer r.wg.Done()
	for {
		e, shutdown := r.knotQueue.Get()
		if !shutdown {
			if err := r.handler(e); err != nil {
				logger.Error("runner", "handler", err)
			}
		} else {
			return
		}
	}
}
func (r *runner) handler(k knot.Interface) error {
	trigger := k.Trigger()
	rawValueAsState := trigger.ToTextWithNoVars(k.State())
	templateNodes := GetTemplates(trigger, k.State())
	for _, tempN := range templateNodes {
		tempN.Execute(rawValueAsState)
		cHashMap := positioner.Default().ContainerOf(tempN)
		if cHashMap == nil {
			continue
		}
		for hash := range cHashMap {
			c := container.NewWithDefaultFetcher(hash)
			if c == nil {
				continue
			}
			ok := c.Forward(tempN, rawValueAsState)
			if !ok {
				continue
			}
			R, S := c.R()
			for nv, rn := range R {
				for _, s := range S[nv] {
					r.knotQueue.Add(knot.New(rn, s))
				}
			}
		}
	}
	return nil
}
func GetTemplates(t node.Interface, state string) m.Map[node.Interface] {
	return m.Map[node.Interface](template.Default().GetTemplate(t, state))
}
func New(knotQueue Queue[knot.Interface], externalQueue Queue[node.Interface]) Runner {
	return &runner{
		workerNum:     5,
		wg:            &sync.WaitGroup{},
		stopChan:      make(chan struct{}),
		knotQueue:     knotQueue,
		externalQueue: externalQueue,
	}
}
