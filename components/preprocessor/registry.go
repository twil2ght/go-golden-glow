package preprocessor

import (
	"goldenglow/components"
	"goldenglow/pkg/log"
	"goldenglow/utils"
)

var (
	logger = log.Default()
)

type registry struct {
	handlers map[string]Handler
}

func (r *registry) Preprocess(message components.Message) string {
	var (
		value = message.Value()
		tag   = message.Tag()
	)
	if handler, ok := r.handlers[tag]; ok {
		return handler(value)
	}
	return value
}

func (r *registry) Register(pluginName, tag string, handler Handler) error {
	if err := utils.NotNull(
		"pluginName", pluginName,
		"tag", tag,
		"handler", handler,
	); err != nil {
		return err
	}
	key := pluginName + ":" + tag
	r.handlers[key] = handler
	logger.Info("Preprocessor: Register preprocessor", "pluginName", pluginName, "tag", tag)
	return nil
}
func NewRegistry() Registry {
	return &registry{
		handlers: make(map[string]Handler),
	}
}
