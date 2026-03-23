package plugins

import (
	"fmt"
	"goldenglow/dataGen"
	"goldenglow/plugin"
)

type registry struct {
	plugins []Item
	exeReg  ExecuteRegistry
	langReg plugin.LangRegistry
	dataGen dataGen.DataGen
}

func (r *registry) Register(plugin Item) error {
	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}
	r.plugins = append(r.plugins, plugin)
	return nil
}

func (r *registry) Init() {
	for _, p := range r.plugins {
		r.parsePlugin(p)
		err := p.Setup()
		if err != nil {
			panic(err)
		}
	}
}

// TODO lang 要从json读取
func (r *registry) parsePlugin(plugin Item) {
	var (
		pluginName      = plugin.Name()
		executorHandler = plugin.ExecuteHandler()
	)
	if err := r.exeReg.Register(pluginName, executorHandler); err != nil {
		panic(pluginName + " " + err.Error())
	}
}
func (r *registry) ShutDown() {
	for _, p := range r.plugins {
		p.Cleanup()
	}
}
func NewRegistry(exeReg ExecuteRegistry, langReg plugin.LangRegistry) Registry {
	return &registry{
		exeReg:  exeReg,
		langReg: langReg,
	}
}
