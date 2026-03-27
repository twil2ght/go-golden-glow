package scheduler

import (
	"context"
	"goldenglow/components/preprocessor"
	"goldenglow/components/queue"
	"goldenglow/components/runner"
	"goldenglow/components/source"
	"goldenglow/node"
	"goldenglow/pkg/log"
	"sync"
)

var logger = log.Default()

type scheduler struct {
	queue       queue.Queue
	mainStream  source.MainStream
	runner      runner.Instance
	processor   preprocessor.Instance
	nodeFactory node.Factory
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func (s *scheduler) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	logger.Info("Scheduler stopped gracefully")
	return nil
}

func (s *scheduler) Start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case message, ok := <-s.mainStream.C():
				if !ok {
					return
				}
				input := s.processor.Preprocess(message)
				if err := s.queue.In(input); err != nil {
					logger.Warn("Failed to queue message", "error", err)
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
					logger.Error("Failed to create item", "error", err)
					continue
				}
				if item == nil {
					logger.Error("Created nil item")
					continue
				}
				if err := s.runner.Run(item); err != nil {
					logger.Error("Failed to run item", "error", err)
				}
			}
		}
	}()

	logger.Info("Scheduler started")
	<-s.ctx.Done()
	return nil
}
func NewScheduler(
	mainStream source.MainStream,
	processor preprocessor.Instance,
	queue queue.Queue,
	runner runner.Instance,
	nodeFactory node.Factory,
) Scheduler {
	return &scheduler{
		mainStream:  mainStream,
		processor:   processor,
		queue:       queue,
		runner:      runner,
		nodeFactory: nodeFactory,
		wg:          sync.WaitGroup{},
	}
}
