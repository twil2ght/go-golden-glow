package plugins

import (
	"goldenglow/components/preprocessor"
	"goldenglow/components/source"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/lang"
)

type Parameters map[string]string
type Item interface {
	Name() string
	OnRegisterExecutor(reg executor.Registry) error
	OnRegisterDataGen(reg dataGen.Registry) error
	OnRegisterLang(reg lang.Registry) error
	OnRegisterInputSource(reg source.Registry) error
	OnRegisterPreprocessor(reg preprocessor.Registry) error
	Setup() error
	Cleanup()
}
type Registry interface {
	Register(plugin Item) error
	Init()
	ShutDown()
	GetAll() []Item
}
