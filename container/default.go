package container

import (
	"goldenglow/node"
	"goldenglow/storage"
	"goldenglow/variable"
)

var (
	db                    = storage.DefaultJSONRepo()
	encoder               = node.DefaultEncoder()
	fetcherInstance, _    = NewFetcher(db, node.DefaultFactory())
	positionerInstance, _ = NewPositioner(db, encoder)
	factoryInstance, _    = NewFactory(fetcherInstance, positionerInstance)
	storeInstance, _      = NewStore(db, encoder)
)

func DefaultFactory() Factory {
	factoryInstance.WithVarReg(variable.VarReg)
	return factoryInstance
}
func DefaultStore() Store { return storeInstance }
