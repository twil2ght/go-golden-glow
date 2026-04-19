package plugin

import "goldenglow/pkg/registry"

var (
	DefaultManager = registry.New[Interface]()
)
