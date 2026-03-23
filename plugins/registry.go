package plugins

import (
	"fmt"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/lang"
)

var (
	globalRegistry = NewRegistry(nil, nil, nil)
)

func Subscribe(plugin Item) error {
	return globalRegistry.Register(plugin)
}

type registry struct {
	plugins []Item
	exeReg  executor.Registry
	langReg lang.Registry
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
func (r *registry) Run() error {
	if err := r.exeReg.RunAll(); err != nil {
		return err
	}
	if err := r.dataGen.RunAll(); err != nil {
		return err
	}
	if err := r.langReg.RunAll(); err != nil {
		return err
	}
	return nil
}

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
func NewRegistry(exeReg executor.Registry, langReg lang.Registry, dataGen dataGen.Registry) Registry {
	return &registry{
		exeReg:  exeReg,
		langReg: langReg,
		dataGen: dataGen,
	}
}
