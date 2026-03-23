package plugins

import (
	"fmt"
	"goldenglow/dataGen"
)

type registry struct {
	plugins []Item
	exeReg  ExecuteRegistry
	langReg LangRegistry
	dataGen dataGen.Registry
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
		pluginName = plugin.Name()
	)
	if err := plugin.OnRegisterDataGen(r.dataGen); err != nil {
		panic(fmt.Sprintf("%s.OnRegisterDataGen err:%v", pluginName, err))
	}
	if err := plugin.OnRegisterLang(r.langReg); err != nil {
		panic(fmt.Sprintf("%s.OnRegisterLang err:%v", pluginName, err))
	}
	if err := plugin.OnRegisterExecutor(r.exeReg); err != nil {
		panic(fmt.Sprintf("%s.OnRegisterExecutor err:%v", pluginName, err))
	}
}
func (r *registry) ShutDown() {
	for _, p := range r.plugins {
		p.Cleanup()
	}
}
func NewRegistry(exeReg ExecuteRegistry, langReg LangRegistry) Registry {
	return &registry{
		exeReg:  exeReg,
		langReg: langReg,
	}
}
