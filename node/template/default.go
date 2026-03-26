package template

import (
	"goldenglow/node"
	"goldenglow/storage"
	"goldenglow/variable"
)

var (
	sourceInstance          = &source{repo: storage.DefaultJSONRepo(), factory: node.DefaultFactory()}
	templateCoreInstance, _ = New(sourceInstance, variable.VarReg)
)

func DefaultCore() Core {
	return templateCoreInstance
}
