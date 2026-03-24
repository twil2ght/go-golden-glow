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
	OnRegisterExecutor(executorRegistry executor.Registry) error
	OnRegisterDataGen(dataGenRegistry dataGen.Registry) error
	OnRegisterLang(langRegistry lang.Registry) error
	OnRegisterInputSource(sourceRegistry source.Registry) error
	OnRegisterPreprocessor(preprocessorRegistry preprocessor.Registry) error
	Setup() error
	Cleanup()
}
type Registry interface {
	Register(plugin Item) error
	Init()
	ShutDown()
}
