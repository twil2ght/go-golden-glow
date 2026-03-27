package scheduler

import (
	"context"
	"goldenglow/components"
	"goldenglow/components/preprocessor"
	"goldenglow/components/queue"
	"goldenglow/components/receiver"
	"goldenglow/components/runner"
	"goldenglow/node"
	"goldenglow/pkg/log"
	"sync"
)

var loggerLite = log.Default()

type schedulerLite struct {
	queue       queue.Queue
	runner      runner.Instance
	processor   preprocessor.Instance
	nodeFactory node.Factory
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	rcv         chan components.Message
}

func (s *schedulerLite) OnRegisterReceiver(reg receiver.Registry) error {
	return reg.Register("main", reg.Subscriptions(), func() chan<- components.Message {
		return s.rcv
	})
}

func (s *schedulerLite) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	loggerLite.Info("Scheduler stopped gracefully")
	return nil
}

func (s *schedulerLite) Start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case message, ok := <-s.rcv:
				loggerLite.Debug("Received message", "message", message)
				if !ok {
					return
				}
				input := s.processor.Preprocess(message)
				if err := s.queue.In(input); err != nil {
					loggerLite.Warn("Failed to queue message", "error", err)
					continue
				}
			}
		}
	}()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case output, ok := <-s.queue.Out():
				if !ok {
					return
				}
				item, err := s.nodeFactory.New(output)
				if err != nil {
					loggerLite.Error("Failed to create item", "error", err)
					continue
				}
				if item == nil {
					loggerLite.Error("Created nil item")
					continue
				}
				if err := s.runner.Run(item); err != nil {
					loggerLite.Error("Failed to run item", "error", err)
				}
			}
		}
	}()

	loggerLite.Info("Scheduler started")
	return nil
}
func NewLiteScheduler(
	processor preprocessor.Instance,
	queue queue.Queue,
	runner runner.Instance,
	nodeFactory node.Factory,
) Scheduler {
	return &schedulerLite{
		processor:   processor,
		queue:       queue,
		runner:      runner,
		nodeFactory: nodeFactory,
		rcv:         make(chan components.Message),
		wg:          sync.WaitGroup{},
	}
}
