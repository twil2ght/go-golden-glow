package checker

import "goldenglow/executor"

type Handler func(parameters executor.Parameters) bool

type Registry interface {
	RunAll() error
	Register(pluginName string, handler Handler) error
}
type RegisterItem interface {
	OnRegisterChecker(reg Registry) error
}
