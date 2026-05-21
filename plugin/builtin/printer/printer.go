package printer

import (
	"goldenglow/m"
	"goldenglow/pkg/datagen"
	"goldenglow/pkg/node/handler"
	"goldenglow/plugin"
)

func init() {
	plugin.DefaultManager.Register(Name, &printer{})
}

type printer struct {
}

func (p *printer) Init() {

}

func (p *printer) Shutdown() {

}

const (
	TypeMsg = "msg"
	Name    = "printer"
)

func (p *printer) OnRegisterDataGen(gen datagen.Generator) {
	provider := datagen.NewProvider()
	provider.Add("print", datagen.NewData(
		[]string{"[Print] $1 @Caller $2"},
		[]string{},
		m.Map[string]{
			TypeMsg: "$1",
		},
		datagen.AsExecutor,
	),
	)
	gen.AddProvider(Name, provider)
}

func (p *printer) OnRegisterExecutor(reg handler.Executor[handler.ExecuteHandler]) {
	reg.Handlers().Register(Name, func(parameters handler.Parameters) {
		msg, _ := parameters.Get(TypeMsg)
		println(msg)
	})
}
