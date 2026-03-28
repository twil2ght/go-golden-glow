package speaker

import (
	"fmt"
	"goldenglow/components/preprocessor"
	"goldenglow/components/source"
	"goldenglow/config"
	"goldenglow/dataGen"
	"goldenglow/executor"
	"goldenglow/lang"
	"goldenglow/pkg/log"
	"goldenglow/plugins"
)

func init() {
	err := plugins.Subscribe(NewSpeakerPlugin())
	if err != nil {
		panic(err)
	}
}

var logger = log.Default()

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
	return fmt.Sprintf("%s says %s to %s", config.GG, message, config.User)
}
func (m *speaker) OnRegisterInputSource(reg source.Registry) error {
	return reg.Register(pluginName, tag, m)
}
func (m *speaker) OnRegisterPreprocessor(reg preprocessor.Registry) error {
	return reg.Register(pluginName, tag, m.preprocess)
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
		logger.Debug("speaker executed", "response", response)
		return nil
	})
}

func (m *speaker) OnRegisterDataGen(reg dataGen.Registry) error {
	var (
		generator = dataGen.NewGenerator(m.Name())
	)
	generator.Add("speak", dataGen.New(
		[]string{fmt.Sprintf("%s should say $1", config.GG)},
		dataGen.Parameters{
			"response": "$1",
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
