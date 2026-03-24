package lang

import (
	"goldenglow/container"
)

var (
	langRegInstance = NewRegistry(container.DefaultStore())
)

func DefaultRegistry() Registry { return langRegInstance }
