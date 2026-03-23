package plugins

import (
	"goldenglow/m"
	"goldenglow/node"
)

type LangItem interface {
	Get() (tv, rv m.Hash)
}
type LangGroup []LangItem
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
	Register(name string, item LangGroup) error
	Init() error
}
type Item interface {
	Name() string
	ExecuteHandler() ExecuteHandler
	Setup() error
	Cleanup()
}
type Registry interface {
	Register(plugin Item) error
	Init()
	ShutDown()
}
