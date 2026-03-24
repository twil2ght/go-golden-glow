package preprocessor

import "goldenglow/components"

type Handler func(msg string) string
type Instance interface {
	Preprocess(message components.Message) string
}
type Registry interface {
	Register(pluginName, tag string, handler Handler) error
}
