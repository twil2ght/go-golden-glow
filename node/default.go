package node

import "goldenglow/variable"

var (
	encoderInstance, _   = NewEncoder(variable.VarReg)
	regulatorInstance, _ = NewRegulator(nil)
	factoryInstance, _   = NewFactory(variable.ToRawText, regulatorInstance)
)

func DefaultEncoder() Encoder {
	return encoderInstance
}
func DefaultFactory() Factory {
	return factoryInstance
}
func DefaultRegulator() Regulator {
	return regulatorInstance
}
