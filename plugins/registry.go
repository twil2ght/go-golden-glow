package plugins

import (
	"fmt"
	"goldenglow/plugin"
)

type registry struct {
	plugins []Item
	exeReg  ExecuteRegistry
	langReg plugin.LangRegistry
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
		p.Init()
	}
}
func (r *registry) parsePlugin(plugin Item) {
	var (
		pluginName      = plugin.Name()
		executorHandler = plugin.ExecuteHandler()
		lang            = plugin.Lang()
	)
	if err := r.exeReg.Register(pluginName, executorHandler); err != nil {
		panic(pluginName + " " + err.Error())
	}
	if err := r.langReg.Register(pluginName, lang); err != nil {
		panic(pluginName + " " + err.Error())
	}
}
func (r *registry) ShutDown() {
	for _, p := range r.plugins {
		p.Shutdown()
	}
}
func NewRegistry(exeReg ExecuteRegistry, langReg plugin.LangRegistry) Registry {
	return &registry{
		exeReg:  exeReg,
		langReg: langReg,
	}
}
