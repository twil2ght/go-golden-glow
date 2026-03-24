package preprocessor

import "errors"

type registry struct {
	handlers map[string]Handler
}

func (r *registry) Register(pluginName, tag string, handler Handler) error {
	if pluginName == "" {
		return errors.New("plugin name is empty")
	}
	if tag == "" {
		return errors.New("plugin tag is empty")
	}
	if handler == nil {
		return errors.New("source is nil")
	}
	key := pluginName + ":" + tag
	r.handlers[key] = handler
	return nil
}
func NewRegistry() Registry {
	return &registry{
		handlers: make(map[string]Handler),
	}
}
