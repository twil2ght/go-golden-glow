package node

import (
	"goldenglow/storage"
	"goldenglow/variable"
)

var (
	encoderInstance, _   = NewEncoder(variable.VarReg)
	regulatorInstance, _ = NewRegulator(storage.DefaultJSONRepo())
	factoryInstance, _   = NewFactory(variable.ToRawText, regulatorInstance)
)

func DefaultEncoder() Encoder {
	return encoderInstance
}
func DefaultFactory() Factory {
	return factoryInstance
}
