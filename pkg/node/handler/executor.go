package handler

import (
	"goldenglow/pkg/node"
	"goldenglow/pkg/registry"
	"goldenglow/variable"
)

type ExecuteHandler func(parameters Parameters)
type CheckHandler func(parameters Parameters) bool
type ExtractorHandler func(parameters Parameters) variable.ValueMap
type Executor struct {
	node.Interface
	handlers registry.Interface[ExecuteHandler]
}

func (e *Executor) Execute(state string) {
	params := GetParameters(e.ToTextWithNoVars(state))
	namespace, _ := params.Get(KeyNamespace)
	handler, _ := e.handlers.Get(namespace)
	if handler != nil {
		handler(params)
	}
}

type Checker struct {
	node.Interface
	handlers registry.Interface[CheckHandler]
}

func (c *Checker) Check(state string) bool {
	params := GetParameters(c.ToTextWithNoVars(state))
	namespace, _ := params.Get(KeyNamespace)
	handler, _ := c.handlers.Get(namespace)
	if handler != nil {
		return handler(params)
	}
	return false
}

type Extractor struct {
	node.Interface
	handlers registry.Interface[ExtractorHandler]
}

func (e *Extractor) KeyDist() string {
	params := GetParameters(e.Value())
	val, _ := params.Get(KeyDist)
	return val
}

func (e *Extractor) Extract(state string) variable.ValueMap {
	params := GetParameters(e.ToTextWithNoVars(state))
	namespace, _ := params.Get(KeyNamespace)
	handler, _ := e.handlers.Get(namespace)
	if handler != nil {
		return handler(params)
	}
	return nil
}
