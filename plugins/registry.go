package plugins

import (
	"fmt"
)

var (
	globalRegistry = NewRegistry()
)

func Subscribe(plugin Item) error {
	return globalRegistry.Register(plugin)
}
func Init()          { globalRegistry.Init() }
func GetAll() []Item { return globalRegistry.GetAll() }

type registry struct {
	plugins []Item
}

func (r *registry) GetAll() []Item {
	return r.plugins
}

// Register TODO 插件的注册需要打印info
func (r *registry) Register(plugin Item) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}
	r.plugins = append(r.plugins, plugin)
	return nil
}

func (r *registry) Init() {
	for _, p := range r.plugins {
		err := p.Setup()
		if err != nil {
			panic(err)
		}
	}
}
func (r *registry) ShutDown() {
	for _, p := range r.plugins {
		p.Cleanup()
	}
}
func NewRegistry() Registry {
	return &registry{}
}
