package source

import "goldenglow/components"

type Source interface {
	C() <-chan components.Message
}
type Registry interface {
	Register(tag string, source Source) error
}
