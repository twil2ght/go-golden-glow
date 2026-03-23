package plugins

type basePlugin struct {
	name string
}

func (b *basePlugin) Cleanup() {}

func (b *basePlugin) Setup() {}

func (b *basePlugin) Name() string {
	return b.name
}

func (b *basePlugin) ExecuteHandler() ExecuteHandler {
	panic(b.name + ":ExecuteHandler not implemented")
}
