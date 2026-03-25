package checker

import "goldenglow/node"

var (
	checkRegInstance = NewRegistry(node.DefaultFactory())
)

func DefaultRegistry() Registry {
	return checkRegInstance
}
