package container

import (
	"goldenglow/node"
)

var (
	fetcherInstance, _    = NewFetcher(nil, node.DefaultFactory())
	positionerInstance, _ = NewPositioner(nil, node.DefaultEncoder())
	factoryInstance, _    = NewFactory(fetcherInstance, positionerInstance)
)

func DefaultFactory() Factory {
	return factoryInstance
}
