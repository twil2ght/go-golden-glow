package source

import "goldenglow/components"

type Source interface {
	C() <-chan string
}
type MainStream interface {
	C() <-chan components.Message
}
type Registry interface {
	Register(pluginName, tag string, source Source) error
}
