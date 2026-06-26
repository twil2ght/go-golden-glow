package runner

import (
	"context"
	"fmt"
	"goldenglow/m"
	"goldenglow/pkg/container"
	"goldenglow/pkg/container/positioner"
	"goldenglow/pkg/knot"
	"goldenglow/pkg/log"
	"goldenglow/pkg/node"
	"goldenglow/pkg/node/template"
	"goldenglow/pkg/registry"
	"goldenglow/pkg/tracer"
	"goldenglow/pkg/workqueue"
	"strings"
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
	IsFinished() bool
}
type runner struct {
	workerNum     int
	wg            *sync.WaitGroup
	stopChan      chan struct{}
	externalQueue Queue[string]
	knotQueue     Queue[knot.Interface]
	nodeFactory   node.Factory
	pending       *atomic.Uint64
	traceMgr      registry.Interface[TraceHandler]
	knotSeq       atomic.Uint64
	knotParents   sync.Map      // map[string]int — key "value\x00state" → parent knotSeq
	idleNotify    chan struct{} // signaled by workers when they finish and system appears idle
}

var logger = log.Default()

func (r *runner) Run(ctx context.Context) {
	r.wg.Add(r.workerNum + 1)

	for i := 0; i < r.workerNum; i++ {
		go r.worker(i)
	}
	go r.watchIdle(ctx, 100*time.Millisecond)

	r.wg.Wait()
}
func (r *runner) watchIdle(ctx context.Context, timeout time.Duration) {
	defer r.wg.Done()

	timer := time.NewTimer(timeout)
	timer.Stop()

	// Kick-start: if system starts idle, start the debounce timer immediately.
	if r.IsFinished() {
		timer.Reset(timeout)
	}

	for {
		select {
		case <-ctx.Done():
			r.knotQueue.Shutdown()
			return
		case <-r.stopChan:
			r.knotQueue.Shutdown()
			return
		case <-r.idleNotify:
			// A worker finished and the system might be idle.
			if r.IsFinished() {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(timeout)
			}
		case <-timer.C:
			// Debounce elapsed; if still idle, trigger the next input cycle.
			if r.IsFinished() {
				r.onIdle()
			}
		}
	}
}
func (r *runner) IsFinished() bool {
	return r.knotQueue.Len() == 0 && r.pending.Load() == 0
}

// onIdle is called when the channel has been empty for the specified duration
func (r *runner) onIdle() {
	r.knotParents.Clear()
	r.nodeFactory.Reset()
	DefaultManager.Range(func(_ string, F IdleHandler) bool {
		return F()
	})
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
		if shutdown {
			return
		}
		r.pending.Add(1)
		func() {
			defer func() {
				if r.pending.Add(^uint64(0)) == 0 && r.knotQueue.Len() == 0 {
					select {
					case r.idleNotify <- struct{}{}:
					default:
					}
				}
			}()
			if err := r.handler(e); err != nil {
				logger.Error("runner", "handler", err)
			}
		}()
	}
}
func (r *runner) handler(k knot.Interface) error {
	trigger := k.Trigger()
	rawValueAsState := trigger.ToTextWithNoVars(k.State())
	knotSeq := int(r.knotSeq.Add(1))

	parentSeq := 0
	if p, ok := r.knotParents.Load(knotKey(trigger.Value(), k.State())); ok {
		parentSeq = p.(int)
	}

	r.traceEvent(tracer.Event{
		Type:          tracer.EventKnotReceived,
		Timestamp:     time.Now(),
		NodeValue:     trigger.Value(),
		State:         k.State(),
		KnotSeq:       knotSeq,
		ParentKnotSeq: parentSeq,
	})

	templateNodes := GetTemplates(trigger, k.State())
	for _, tempN := range templateNodes {
		tempN.Execute(rawValueAsState)
		if len(tempN.VarKeys()) == 0 {
			tempN.Activate()
		}

		cHashMap := positioner.Default().ContainerOf(tempN)
		if cHashMap == nil {
			r.traceEvent(tracer.Event{
				Type:      tracer.EventTemplateSkipped,
				Timestamp: time.Now(),
				NodeValue: tempN.Value(),
				State:     rawValueAsState,
				Detail:    map[string]string{"reason": "no_container"},
				KnotSeq:   knotSeq,
			})
			continue
		}

		r.traceEvent(tracer.Event{
			Type:      tracer.EventTemplateMatched,
			Timestamp: time.Now(),
			NodeValue: tempN.Value(),
			State:     rawValueAsState,
			KnotSeq:   knotSeq,
		})

		for hash := range cHashMap {
			r.traceEvent(tracer.Event{
				Type:      tracer.EventContainerFound,
				Timestamp: time.Now(),
				NodeValue: hash,
				State:     tempN.Value(),
				Detail:    map[string]string{"container_hash": hash},
				KnotSeq:   knotSeq,
			})

			c := container.NewWithDefaultFetcher(hash)
			if c == nil {
				continue
			}
			ok := c.Forward(tempN, rawValueAsState)
			tNodes := tNodeValues(c.T())
			if !ok {
				r.traceEvent(tracer.Event{
					Type:      tracer.EventContainerReject,
					Timestamp: time.Now(),
					NodeValue: hash,
					State:     rawValueAsState,
					Detail:    map[string]string{"container_hash": hash, "t_nodes": tNodes},
					KnotSeq:   knotSeq,
				})
				continue
			}
			r.traceEvent(tracer.Event{
				Type:      tracer.EventContainerForward,
				Timestamp: time.Now(),
				NodeValue: hash,
				State:     rawValueAsState,
				Detail:    map[string]string{"container_hash": hash, "t_nodes": tNodes},
				KnotSeq:   knotSeq,
			})

			R, S := c.R()
			for nv, rn := range R {
				for _, s := range S[nv] {
					key := knotKey(rn.Value(), s)
					//log.Default().Debug("R", "value", rn.ToTextWithNoVars(s), "raw", rn.Value())
					if _, loaded := r.knotParents.LoadOrStore(key, knotSeq); loaded {
						continue
					}
					log.Default().Debug("R", "value", rn.ToTextWithNoVars(s), "raw", rn.Value())
					r.traceEvent(tracer.Event{
						Type:      tracer.EventResultProduced,
						Timestamp: time.Now(),
						NodeValue: rn.Value(),
						State:     s,
						Detail:    map[string]string{"result_raw": rn.ToTextWithNoVars(s)},
						KnotSeq:   knotSeq,
					})
					r.knotQueue.Add(knot.New(rn, s))
				}
			}
		}
	}
	return nil
}

func knotKey(value, state string) string {
	return value + "\x00" + state
}

func tNodeValues(T m.Map[node.Interface]) string {
	vals := make([]string, 0, len(T))
	for _, tn := range T {
		vals = append(vals, tn.Value())
	}
	return strings.Join(vals, "\n")
}
func GetTemplates(t node.Interface, state string) m.Map[node.Interface] {
	return m.Map[node.Interface](template.Default().GetTemplate(t, state))
}
func (r *runner) traceEvent(e tracer.Event) {
	r.traceMgr.Range(func(_ string, h TraceHandler) bool {
		h(e)
		return true
	})
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
		traceMgr:      DefaultTraceManager,
		idleNotify:    make(chan struct{}, 1),
	}
}
func printContainer(T, R m.Map[node.Interface]) {
	for _, t := range T {
		log.Default().Debug(
			"T:",
			"value", t.Value(),
			"state", t.VarSetRegistry().Len(),
			"varKeys", t.VarKeys(),
			"type", fmt.Sprintf("%T", t),
		)
	}
	for _, r := range R {
		log.Default().Debug("R:", "value", r.Value())
	}
}
