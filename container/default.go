package container

import (
	"goldenglow/node"
	"goldenglow/storage"
)

var (
	db                    = storage.DefaultJSONRepo()
	encoder               = node.DefaultEncoder()
	fetcherInstance, _    = NewFetcher(db, node.DefaultFactory())
	positionerInstance, _ = NewPositioner(db, encoder)
	factoryInstance, _    = NewFactory(fetcherInstance, positionerInstance)
	storeInstance         = NewStore(db, encoder)
)

func DefaultFactory() Factory {
	return factoryInstance
}
func DefaultStore() Store { return storeInstance }
