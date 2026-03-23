package dataGen

import (
	"errors"
	"fmt"
	"sync"
)

// TODO 所有功能统一由主的插件注册器来调度，因此不需要全局的dataGen了，让dataGen作为插件注册器的字段
var (
	globalRegistry DataGen
	once           sync.Once
)

type dataGen struct {
	genRegistries map[string]Generator
}

func (d *dataGen) AddGenerator(pluginName string, generator Generator) error {
	if d.genRegistries == nil {
		d.genRegistries = make(map[string]Generator)
	}
	if pluginName == "" {
		return errors.New("pluginName is empty")
	}
	if generator == nil {
		return errors.New("generator is nil")
	}
	d.genRegistries[pluginName] = generator
	return nil
}

func (d *dataGen) RunAll() error {
	for pluginName, generator := range d.genRegistries {
		err := generator.Run()
		if err != nil {
			return fmt.Errorf("plugin %s generator err: %v", pluginName, err)
		}
	}
	return nil
}

func initGlobal() {
	once.Do(func() {
		globalRegistry = NewDataGen()
	})
}
func NewDataGen() DataGen {
	return &dataGen{
		genRegistries: make(map[string]Generator),
	}
}
func AutoRegister(name string, generator Generator) {
	initGlobal()
	err := globalRegistry.AddGenerator(name, generator)
	if err != nil {
		panic(err)
	}
}

func RunAll() error {
	initGlobal()
	return globalRegistry.RunAll()
}
