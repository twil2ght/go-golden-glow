package handler

import (
	"goldenglow/pkg/node"
)

func init() {
	creatorRegistry := node.DefaultFactory.CreatorRegistry()
	creatorRegistry.Register(NodeExecutor, ExecutorCreator)
	creatorRegistry.Register(NodeChecker, CheckerCreator)
	creatorRegistry.Register(NodeExtractor, ExtractorCreator)
}

func ExecutorCreator(value string) node.Interface {
	return &executor{
		Base: New[ExecuteHandler](value),
	}
}
func CheckerCreator(value string) node.Interface {
	return &checker{
		Base: New[CheckHandler](value),
	}
}
func ExtractorCreator(value string) node.Interface {
	return &extractor{
		Base: New[ExtractorHandler](value),
	}
}
