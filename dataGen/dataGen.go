package dataGen

import (
	"errors"
	"fmt"
	"goldenglow/pkg/log"
	"goldenglow/utils"
)

var (
	logger  = log.Default()
	RootDir = utils.RootDir
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
	logger.Info("add generator successfully")
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

func NewDataGen() Registry {
	return &dataGen{
		genRegistries: make(map[string]Generator),
	}
}
