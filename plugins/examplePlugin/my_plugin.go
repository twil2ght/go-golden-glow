package exampleplugin

import (
	"fmt"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/lang"
	"goldenglow/plugins"
)

func init() {
	err := plugins.Subscribe(NewExamplePlugin())
	if err != nil {
		panic(err)
	}
}

const (
	pluginName = "example"
)

type examplePlugin struct {
	plugins.Base
}

func (e *examplePlugin) Name() string {
	return pluginName
}

func (e *examplePlugin) OnRegisterExecutor(reg executor.Registry) error {
	return reg.Register(e.Name(), func(params executor.Parameters) error {
		response := params["response"]
		if response == "" {
			return fmt.Errorf("example plugin: not found response parameter")
		}

		fmt.Printf("[ExamplePlugin] got response: %s\n", response)
		return nil
	})
}

func (e *examplePlugin) OnRegisterDataGen(reg dataGen.Registry) error {
	generator := dataGen.NewGenerator(e.Name())

	generator.Add("Hello", dataGen.New(
		[]string{"Hello $1"},
		dataGen.Parameters{
			"response": "$1",
		},
		dataGen.LangTypeDefault,
	))

	return reg.AddGenerator(e.Name(), generator)
}

func (e *examplePlugin) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(pluginName)
}

func NewExamplePlugin() plugins.Item {
	return &examplePlugin{}
}
