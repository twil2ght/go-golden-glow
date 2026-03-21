package pack

import (
	"fmt"
	"goldenglow/container"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/plugin"
)

const (
	TagIn      = "[in]"
	TagClose   = "[close]"
	ToPack     = "[pack]"
	ToBackpack = "[backpack]"
)

var packer *Engine

type Engine struct {
	pack     []string
	backpack []string
	store    container.Store
}

func (p *Engine) In(val, tag, to string) {
	switch to {
	case ToPack:
		p.handlePack(val, tag)
	case ToBackpack:
		p.handleBackpack(val, tag)
	}
}

func (p *Engine) handlePack(val, tag string) {
	p.pack = append(p.pack, val)
	if tag == TagClose {
		p.closePack()
	}
}

func (p *Engine) handleBackpack(val, tag string) {
	p.backpack = append(p.backpack, val)
	if tag == TagClose {
		p.closeBackpack()
	}
}

func (p *Engine) closePack() {
	if len(p.pack) < 1 {
		p.pack = nil
		return
	}
	condit := p.pack[len(p.pack)-1]
	res := p.pack[:len(p.pack)-1]

	p.store.Do(m.ToHash([]string{condit}), m.ToHash(res))
	p.pack = nil
}

func (p *Engine) closeBackpack() {
	if len(p.backpack) < 1 {
		p.backpack = nil
		return
	}
	res := p.backpack[len(p.backpack)-1]
	condit := p.backpack[:len(p.backpack)-1]

	p.store.Do(m.ToHash(condit), m.ToHash([]string{res}))
	p.backpack = nil
}

// [pack] val & tag(event) & to & user
var (
	head            = "[Pack]"
	Value           = "val"
	Tag             = plugin.Event
	To              = "to"
	User            = "user"
	parameterConfig = &plugin.ParamCfg{
		Head:        head,
		ParamLength: 4,
	}
	parameterKeys = map[string]int{
		Value: 0,
		Tag:   1,
		To:    2,
		User:  3,
	}
	DefaultHandler = func(ctx plugin.HandlerCtx) {
		val, ok := ctx.Get()[Value].(string)
		if !ok {
			return
		}
		tag, ok := ctx.Get()[Tag].(string)
		if !ok {
			return
		}
		to, ok := ctx.Get()[To].(string)
		if !ok {
			return
		}
		packer.In(val, tag, to)
	}
	Handlers = map[string]plugin.Handler{
		TagIn:    DefaultHandler,
		TagClose: DefaultHandler,
	}
	LangAPI = plugin.TRGroup{
		"In_pack": {
			TV: plugin.NV{"[say] [0x01] to [GG] : [Pack_in] [0x02]"},
			RV: plugin.NV{fmt.Sprintf("%s [0x02] & %s & %s & [0x01]", parameterConfig.Head, TagIn, ToPack)},
		},
		"Close_pack": {
			TV: plugin.NV{"[say] [0x01] to [GG] : [Pack_close] [0x02]"},
			RV: plugin.NV{fmt.Sprintf("%s [0x02] & %s & %s & [0x01]", parameterConfig.Head, TagClose, ToPack)},
		},
		"In_backpack": {
			TV: plugin.NV{"[say] [0x01] to [GG] : [Backpack_in] [0x02]"},
			RV: plugin.NV{fmt.Sprintf("%s [0x02] & %s & %s & [0x01]", parameterConfig.Head, TagIn, ToBackpack)},
		},
		"Close_backpack": {
			TV: plugin.NV{"[say] [0x01] to [GG] : [Backpack_close] [0x02]"},
			RV: plugin.NV{fmt.Sprintf("%s [0x02] & %s & %s & [0x01]", parameterConfig.Head, TagClose, ToBackpack)},
		},
	}
)

type PackNode struct {
	node.Base
}

func (p *PackNode) Execute() error {
	return plugin.Execute(
		parameterConfig.Head,
		p.Parse,
		&plugin.HandlerCtxBase{},
		Handlers,
	)
}
func (p *PackNode) Parse() (m.H, error) {
	return plugin.ParseFunc(p.Base, parameterConfig, parameterKeys)
}
func NewPackNode(b node.Base) node.Item {
	return &PackNode{Base: b}
}
func New(p *Engine) plugin.Item {
	if p != nil {
		packer = p
	}
	return &plugin.Base{
		ParamCfg:    parameterConfig,
		NodeCreator: NewPackNode,
		LangAPI:     LangAPI,
		Name:        head,
	}
}
func NewEngine(store container.Store) *Engine {
	return &Engine{
		store: store,
	}
}
