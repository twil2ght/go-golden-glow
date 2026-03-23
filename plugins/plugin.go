package plugins

import (
	"goldenglow/plugin"
)

type basePlugin struct {
	name string
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
