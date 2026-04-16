package container

import (
	"goldenglow/node"
	"goldenglow/storage"
	"goldenglow/variable"
)

var (
	db                    = storage.DefaultJSONRepo()
	encoder               = node.DefaultEncoder()
	nFactory              = node.DefaultFactory()
	fetcherInstance, _    = NewFetcher(db, nFactory)
	positionerInstance, _ = NewPositioner(db, encoder)
	factoryInstance, _    = NewFactory(fetcherInstance, positionerInstance, nFactory)
	storeInstance, _      = NewStore(db, encoder)
)

func DefaultFactory() Factory {
	factoryInstance.WithVarReg(variable.VarReg)
	return factoryInstance
}
func DefaultPositioner() Positioner {
	return positionerInstance
}
func DefaultStore() Store { return storeInstance }
