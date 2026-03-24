package source

type Source interface {
	C() <-chan string
}
type Registry interface {
	Register(tag string, source Source) error
}
