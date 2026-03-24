package source

import (
	"errors"
	"goldenglow/components"
)

type registry struct {
	sources map[string]Source
}

func (r *registry) C() <-chan components.Message {
	//TODO implement me
	panic("implement me")
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
