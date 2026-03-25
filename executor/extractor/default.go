package extractor

import "goldenglow/node"

var (
	extractorRegInstance = NewRegistry(node.DefaultFactory())
)

func DefaultRegistry() Registry {
	return extractorRegInstance
}
