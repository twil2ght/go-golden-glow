package plugins

import (
	"fmt"
	"goldenglow/pkg/log"
)

var (
	globalRegistry = NewRegistry()
	logger         = log.Default()
)

func Subscribe(plugin Item) error {
	return globalRegistry.Register(plugin)
}

func Init()          { globalRegistry.Init() }
func GetAll() []Item { return globalRegistry.GetAll() }
func ShutDown()      { globalRegistry.ShutDown() }

type registry struct {
	plugins []Item
}

func (r *registry) GetAll() []Item {
	return r.plugins
}

func (r *registry) Register(plugin Item) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}

	r.plugins = append(r.plugins, plugin)

	logger.Info(fmt.Sprintf("plugin registered: [%s]", plugin.Name()))

	return nil
}

func (r *registry) Init() {
	total := len(r.plugins)
	logger.Info(fmt.Sprintf("start initializing all plugins, total count: %d", total))

	for _, p := range r.plugins {
		err := p.Setup()
		if err != nil {
			panic(err)
		}
		logger.Info(fmt.Sprintf("plugin [%s] init success", p.Name()))
	}
}

func (r *registry) ShutDown() {
	for _, p := range r.plugins {
		p.Cleanup()
	}
	logger.Info("all plugins shutdown completed")
}

func NewRegistry() Registry {
	return &registry{}
}
