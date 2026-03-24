package ExamplePlugin

import (
	"fmt"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/plugins"
)

func init() {
	err := plugins.Subscribe(NewMyPlugin())
	if err != nil {
		panic(err)
	}
}

type myPlugin struct {
	plugins.Base
}

func (m *myPlugin) OnRegisterExecutor(reg executor.Registry) error {
	return reg.Register(m.Name(), func(params executor.Parameters) error {
		var (
			response = params["response"]
		)
		if response == "" {
			return fmt.Errorf("not found response")
		}
		fmt.Println(response)
		return nil
	})
}

func (m *myPlugin) OnRegisterDataGen(reg dataGen.Registry) error {
	var (
		generator = dataGen.NewGenerator(m.Name())
	)
	generator.Add("Hello", dataGen.New(
		[]string{"Hello $1"},
		dataGen.Parameters{
			"response": "World",
		},
		dataGen.LangTypeDefault,
	))
	return reg.AddGenerator(m.Name(), generator)
}

func NewMyPlugin() plugins.Item {
	return &myPlugin{}
}
