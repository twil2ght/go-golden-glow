package node

import (
	"goldenglow/variable"
)

var (
	encoderInstance, _ = NewEncoder(variable.VarReg)
	factoryInstance, _ = NewFactory(variable.ToRawText)
)

func DefaultEncoder() Encoder {
	return encoderInstance
}
func DefaultFactory() Factory {
	return factoryInstance
}
