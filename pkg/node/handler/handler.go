package handler

import (
	"goldenglow/pkg/node"
	"goldenglow/pkg/registry"
	"goldenglow/variable"
)

type Executor[T any] interface {
	Handlers() registry.Interface[T]
	OnRegisterFactory(factory node.Factory)
}
type ExecuteHandler func(parameters Parameters)
type CheckHandler func(parameters Parameters) bool
type ExtractorHandler func(parameters Parameters) variable.ValueMap

type Base[T any] struct {
	node.Interface
	handlers registry.Interface[T]
}

func New[T any](value string, handlers registry.Interface[T]) Base[T] {
	return Base[T]{
		Interface: node.New(value),
		handlers:  handlers,
	}
}
func (b *Base[T]) Handlers() registry.Interface[T] {
	return b.handlers
}

type executor struct {
	Base[ExecuteHandler]
}

func (e *executor) OnRegisterFactory(factory node.Factory) {
	factory.CreatorRegistry().Register(NodeExecutor, func(value string) node.Interface {
		return &executor{
			Base: New[ExecuteHandler](value, e.handlers),
		}
	})
}

func (e *executor) Execute(state string) {
	params := GetParameters(e.ToTextWithNoVars(state))
	namespace, _ := params.Get(KeyNamespace)
	handler, _ := e.handlers.Get(namespace)
	if handler != nil {
		handler(params)
	}
}
func NewExecutor() Executor[ExecuteHandler] {
	return &executor{
		Base: New("", registry.New[ExecuteHandler]()),
	}
}

type checker struct {
	Base[CheckHandler]
}

func (c *checker) OnRegisterFactory(factory node.Factory) {
	factory.CreatorRegistry().Register(NodeExecutor, func(value string) node.Interface {
		return &checker{
			Base: New[CheckHandler](value, c.handlers),
		}
	})
}

func (c *checker) Check(state string) bool {
	params := GetParameters(c.ToTextWithNoVars(state))
	namespace, _ := params.Get(KeyNamespace)
	handler, _ := c.handlers.Get(namespace)
	if handler != nil {
		return handler(params)
	}
	return false
}
func NewChecker() Executor[CheckHandler] {
	return &checker{
		Base: New("", registry.New[CheckHandler]()),
	}
}

type extractor struct {
	Base[ExtractorHandler]
}

func (e *extractor) OnRegisterFactory(factory node.Factory) {
	factory.CreatorRegistry().Register(NodeExecutor, func(value string) node.Interface {
		return &extractor{
			Base: New[ExtractorHandler](value, e.handlers),
		}
	})
}

func (e *extractor) KeyDist() string {
	params := GetParameters(e.Value())
	val, _ := params.Get(KeyDist)
	return val
}

func (e *extractor) Extract(state string) variable.ValueMap {
	params := GetParameters(e.ToTextWithNoVars(state))
	namespace, _ := params.Get(KeyNamespace)
	handler, _ := e.handlers.Get(namespace)
	if handler != nil {
		return handler(params)
	}
	return nil
}
func NewExtractor() Executor[ExtractorHandler] {
	return &extractor{
		Base: New("", registry.New[ExtractorHandler]()),
	}
}
