package speaker

import (
	"fmt"
	"goldenglow/components/preprocessor"
	"goldenglow/components/source"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/lang"
	"goldenglow/plugins"
)

func init() {
	err := plugins.Subscribe(NewSpeakerPlugin())
	if err != nil {
		panic(err)
	}
}

const (
	tag        = "self"
	pluginName = "speaker"
)

type speaker struct {
	plugins.Base
	responseChan chan string
}

func (m *speaker) OnRegisterLang(reg lang.Registry) error {
	return reg.Register(pluginName)
}

func (m *speaker) C() <-chan string {
	return m.responseChan
}
func (m *speaker) preprocess(message string) string {
	//TODO implement me
	return ""
}
func (m *speaker) OnRegisterInputSource(reg source.Registry) error {
	return reg.Register(tag, m)
}
func (m *speaker) OnRegisterPreprocessor(reg preprocessor.Registry) error {
	return reg.Register(tag, m.preprocess)
}
func (m *speaker) OnRegisterExecutor(reg executor.Registry) error {
	return reg.Register(m.Name(), func(params executor.Parameters) error {
		var (
			response = params["response"]
		)
		if response == "" {
			return fmt.Errorf("not found response")
		}
		m.responseChan <- response
		return nil
	})
}

func (m *speaker) OnRegisterDataGen(reg dataGen.Registry) error {
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
func (m *speaker) Name() string {
	return pluginName
}
func NewSpeakerPlugin() plugins.Item {
	return &speaker{
		responseChan: make(chan string, 10),
	}
}
