package plugins

import (
	"goldenglow/node"
	"goldenglow/plugin"
)

type basePlugin struct {
	name        string
	langAPI     plugin.TRGroup
	defaultNode node.Item
	paramLen    int
}

func (b *basePlugin) Shutdown() {}

func (b *basePlugin) Init() {}

func (b *basePlugin) Name() string {
	return b.name
}

func (b *basePlugin) Lang() plugin.TRGroup {
	//TODO implement me
	panic("implement me")
}

func (b *basePlugin) ExecuteHandler() ExecuteHandler {
	//TODO implement me
	panic("implement me")
}
