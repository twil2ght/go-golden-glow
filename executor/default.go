package executor

import "goldenglow/node"

var (
	exeRegInstance = NewRegistry(node.DefaultFactory())
)

func DefaultRegistry() Registry {
	return exeRegInstance
}
