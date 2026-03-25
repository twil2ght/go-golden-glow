package extractor

import (
	"goldenglow/executor"
	"goldenglow/variable"
)

type Handler func(params executor.Parameters) (variable.Item, error)
type Registry interface {
	Register(pluginName string, handler Handler) error
	RunAll() error
}
type RegisterItem interface {
	OnRegisterExtractor(reg Registry) error
}
