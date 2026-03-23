package executor

type Parameters map[string]string
type Handler func(params Parameters) error

type Registry interface {
	RunAll() error
	Register(name string, method Handler) error
}
