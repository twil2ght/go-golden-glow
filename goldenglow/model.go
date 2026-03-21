package goldenglow

import (
	"goldenglow/node"
	"goldenglow/pkg/log"
	"goldenglow/runner"
)

type Source interface {
	C() <-chan Message
}
type Queue interface {
	In(string) error
	Out() <-chan string
}
type ProcessorHandler func(msg string) string
type Processor interface {
	Do(message Message) string
	Register(tag string, handler ProcessorHandler) error
}
type Message interface {
	Value() string
	Tag() string
}
type scheduler struct {
	queue       Queue
	source      Source
	logger      log.Logger
	runner      runner.Instance
	processor   Processor
	nodeFactory node.Factory
}

func (s *scheduler) Start() {
	go func() {
		for message := range s.source.C() {
			input := s.processor.Do(message)
			err := s.queue.In(input)
			if err != nil {
				s.logger.Warn("Failed to queue message", "error", err)
				break
			}
		}
	}()
	go func() {
		for output := range s.queue.Out() {
			item, err := s.nodeFactory.New(output)
			if err != nil {
				s.logger.Warn("Failed to create item", "error", err)
			}
			err = s.runner.Run(item)
			if err != nil {
				s.logger.Warn("Failed to run item", "error", err)
			}
		}
	}()
	select {}
}
