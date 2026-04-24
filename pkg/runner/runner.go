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
	"goldenglow/pkg/workqueue"
	"sync"
	"sync/atomic"
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
	externalQueue Queue[string]
	knotQueue     Queue[knot.Interface]
	nodeFactory   node.Factory
	pending       *atomic.Uint64
}

var logger = log.Default()

func (r *runner) Run(ctx context.Context) {
	r.wg.Add(r.workerNum + 1)

	for i := 0; i < r.workerNum; i++ {
		go r.worker(i)
	}
	go r.watchIdle(ctx, 10*time.Millisecond, 100*time.Millisecond)

	r.wg.Wait()
}
func (r *runner) watchIdle(ctx context.Context, checkInterval time.Duration, timeout time.Duration) {
	defer r.wg.Done()
	ticker := time.NewTicker(checkInterval)
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
			if r.IsFinished() {
				if !isIdle {
					// Just became idle
					isIdle = true
					idleStartTime = time.Now()
				}

				// Check if idle time has exceeded threshold
				if time.Since(idleStartTime) > timeout {
					r.onIdle()
					isIdle = false
				}
			} else {
				// Channel is not empty, reset idle state
				isIdle = false
			}
		}
	}
}
func (r *runner) IsFinished() bool {
	return r.knotQueue.Len() == 0 && r.pending.Load() == 0
}

// onIdle is called when the channel has been empty for the specified duration
func (r *runner) onIdle() {
	r.nodeFactory.Reset()
	n, shutdown := r.externalQueue.Get()
	if !shutdown {
		initKnot := knot.New(r.nodeFactory.Create(n), "")
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
			r.pending.Add(1)
			func() {
				defer r.pending.Add(^uint64(0))
				if err := r.handler(e); err != nil {
					logger.Error("runner", "handler", err)
				}
			}()
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
		if len(tempN.VarKeys()) == 0 {
			tempN.Activate()
		}

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
			//printContainer(c.T(), make(m.Map[node.Interface]))
			//log.Default().Debug("cut")
			if !ok {
				continue
			}
			R, S := c.R()

			//printContainer(make(m.Map[node.Interface]), R)
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
func New(workNum int, externalQueue Queue[string], nodeFactory node.Factory) Runner {
	return &runner{
		workerNum:     workNum,
		wg:            &sync.WaitGroup{},
		stopChan:      make(chan struct{}),
		knotQueue:     workqueue.New[knot.Interface](),
		externalQueue: externalQueue,
		nodeFactory:   nodeFactory,
		pending:       new(atomic.Uint64),
	}
}
func printContainer(T, R m.Map[node.Interface]) {
	for _, t := range T {
		log.Default().Debug("T:", "value", t.Value(), "state", t.VarSetRegistry().Len())
	}
	for _, r := range R {
		log.Default().Debug("R:", "value", r.Value())
	}
}
