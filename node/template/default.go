package template

import "goldenglow/variable"

var (
	sourceInstance          = &source{}
	templateCoreInstance, _ = New(sourceInstance, variable.VarReg)
)

func DefaultCore() Core {
	return templateCoreInstance
}
