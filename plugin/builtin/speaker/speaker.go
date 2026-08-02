package speaker

import (
	"fmt"
	"goldenglow/config"
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/log"
	"goldenglow/pkg/messageQueue"
	"goldenglow/pkg/node/handler"
	"goldenglow/plugin"
)

func init() {
	plugin.DefaultManager.Register(name, New())
}

const (
	name = "speaker"

	KeyMsg = "response"
	KeyTo  = "to"
)

type speaker struct {
	responseChan chan string
}

func (m *speaker) OnRegisterMsgProvider(reg messageQueue.Manager) {
	reg.Add(name, m.responseChan)
}

func (m *speaker) decorate(msg, to string) string {
	if to == "" {
		return fmt.Sprintf("%s says : %s", config.GG, msg)
	}
	return fmt.Sprintf("%s says to %s : %s", config.GG, to, msg)
}

func (m *speaker) OnRegisterExecutor(reg handler.Executor[handler.ExecuteHandler]) {
	reg.Handlers().Register(name, func(parameters handler.Parameters) {
		var (
			msg, _ = parameters.Get(KeyMsg)
			to, _  = parameters.Get(KeyTo)
		)
		if msg == "" {
			return
		}
		log.Default().Info("speaking", "msg", msg, "to", to)
		m.responseChan <- m.decorate(msg, to)
	})
}

func (m *speaker) OnRegisterDataGen(gen datagen.Generator) {
	provider := datagen.NewProvider()
	provider.Add("speak", datagen.NewData(
		[]string{"[Speak] $1 -> $2"},
		[]string{},
		map[string]string{
			KeyMsg: "$1",
			KeyTo:  "$2",
		},
		datagen.AsExecutor,
	))
	provider.Add("speakWithoutTarget", datagen.NewData(
		[]string{"[Speak] $1"},
		[]string{},
		map[string]string{
			KeyMsg: "$1",
			KeyTo:  "",
		},
		datagen.AsExecutor,
	))
	gen.AddProvider(name, provider)
}

func (m *speaker) Init() {}

func (m *speaker) Shutdown() {}

func New() plugin.Interface {
	return &speaker{
		responseChan: make(chan string, 10),
	}
}
