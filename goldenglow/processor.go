package goldenglow

import "errors"

type processor struct {
	registries map[string]ProcessorHandler
}

func (p *processor) Do(message Message) string {
	if handler, ok := p.registries[message.Tag()]; ok {
		return handler(message.Value())
	}
	return message.Value()
}

func (p *processor) Register(tag string, handler ProcessorHandler) error {
	if handler == nil {
		return errors.New("nil processorHandler")
	}
	if tag == "" {
		return errors.New("empty tag")
	}
	p.registries[tag] = handler
	return nil
}
func NewProcessor() Processor {
	return &processor{
		registries: make(map[string]ProcessorHandler),
	}
}
