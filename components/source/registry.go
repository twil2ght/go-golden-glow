package source

import (
	"errors"
	"goldenglow/components"
)

type registry struct {
	sources map[string]Source
}

func (r *registry) C() <-chan components.Message {
	mainstream := make(chan components.Message, 10)

	for tag, ch := range r.sources {
		go func(source Source, tag string) {
			for msg := range source.C() {
				mainstream <- components.NewMsg(msg, tag)
			}
		}(ch, tag)
	}

	return mainstream
}

func (r *registry) Register(pluginName, tag string, source Source) error {
	if pluginName == "" {
		return errors.New("plugin name is empty")
	}
	if tag == "" {
		return errors.New("plugin tag is empty")
	}
	if source == nil {
		return errors.New("source is nil")
	}
	key := pluginName + ":" + tag
	r.sources[key] = source
	return nil
}
func NewRegistry() Registry {
	return &registry{
		sources: make(map[string]Source),
	}
}
