package plugins

import (
	"goldenglow/dataGen"
	"goldenglow/m"
	"goldenglow/node"
)

type Parameters map[string]string
type ExecuteHandler func(params Parameters) error

type ExecuteRegistry interface {
	Init() (string, node.Creator)
	Register(name string, method ExecuteHandler) error
}
type LangRepo interface {
	Save(tv, rv m.Hash) error
}
type LangRegistry interface {
	Register(pluginName string) error //传pluginName去找data路径
	RunAll() error
}
type Item interface {
	Name() string
	OnRegisterExecutor(executorRegistry ExecuteRegistry) error
	OnRegisterDataGen(dataGenRegistry dataGen.Registry) error
	OnRegisterLang(langRegistry LangRegistry) error
	Setup() error
	Cleanup()
}
type Registry interface {
	Register(plugin Item) error
	Init()
	ShutDown()
}
