package plugins

import (
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/lang"
)

type Parameters map[string]string
type Item interface {
	Name() string
	OnRegisterExecutor(executorRegistry executor.Registry) error
	OnRegisterDataGen(dataGenRegistry dataGen.Registry) error
	OnRegisterLang(langRegistry lang.Registry) error
	Setup() error
	Cleanup()
}
type Registry interface {
	Register(plugin Item) error
	Init()
	ShutDown()
}
