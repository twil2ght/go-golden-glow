package template

import (
	"goldenglow/container"
	"goldenglow/node"
	"goldenglow/storage"
	"goldenglow/variable"
)

var (
	sourceInstance = &source{
		repo:       storage.DefaultJSONRepo(),
		factory:    node.DefaultFactory(),
		positioner: container.DefaultPositioner(),
	}
	templateCoreInstance, _ = New(sourceInstance, variable.VarReg)
)

func DefaultCore() Core {
	return templateCoreInstance
}
