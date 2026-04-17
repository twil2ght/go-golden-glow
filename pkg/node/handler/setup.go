package handler

import (
	"goldenglow/pkg/node"
	"goldenglow/pkg/registry"
)

func init() {
	creatorRegistry := node.DefaultFactory().CreatorRegistry()
	creatorRegistry.Register(NodeExecutor, ExecutorCreator)
	creatorRegistry.Register(NodeChecker, CheckerCreator)
	creatorRegistry.Register(NodeExtractor, ExtractorCreator)
}

func ExecutorCreator(value string) node.Interface {
	return &Executor{
		Interface: node.New(value),
		handlers:  registry.New[ExecuteHandler](),
	}
}
func CheckerCreator(value string) node.Interface {
	return &Checker{
		Interface: node.New(value),
		handlers:  registry.New[CheckHandler](),
	}
}
func ExtractorCreator(value string) node.Interface {
	return &Extractor{
		Interface: node.New(value),
		handlers:  registry.New[ExtractorHandler](),
	}
}
