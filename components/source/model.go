package source

type Source interface {
	C() <-chan string
}
type Registry interface {
	Register(pluginName, tag string, source Source) error
}
