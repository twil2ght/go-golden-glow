package kv

import (
	"fmt"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/pkg/log"
	"goldenglow/plugin"
	"goldenglow/variable"
)

type dataBarn interface {
	Set(string, string) error
	Get(string) string
	Del(string, string) error
}
type Engine struct {
	barn dataBarn
}
type KVNode struct {
	node.Base
}

func (e *Engine) Add(k, v string) {
	_ = e.barn.Set(k, v)
}
func (e *Engine) Del(k, v string) {
	_ = e.barn.Del(k, v)
}
func (e *Engine) Get(k string) string {
	return e.barn.Get(k)
}
func (e *Engine) KV(param m.H) (string, string) {
	k, ok := param[k].(string)
	if !ok {
		return "", ""
	}
	v, ok := param[v].(string)
	if !ok {
		return "", ""
	}
	return k, v
}

// [KV] k & v & [add]([del])
var (
	head        = "[KV]"
	paramConfig = &plugin.ParamCfg{
		Head:        head,
		ParamLength: 3,
	}
	eventAdd  = "[add]"
	eventDel  = "[Del]"
	eventGet  = "[HGet]"
	eventPush = "[Push]"
	k         = "k"
	v         = "v"
	paramKeys = map[string]int{
		k:            0,
		v:            1,
		plugin.Event: 2,
	}
	paramHandlers = map[string]plugin.Handler{
		eventAdd: func(ctx plugin.HandlerCtx) {
			k, v := engine.KV(ctx.Get())
			if k != "" && v != "" {
				engine.Add(k, v)
			}
		},
		eventDel: func(ctx plugin.HandlerCtx) {
			k, v := engine.KV(ctx.Get())
			if k != "" && v != "" {
				engine.Del(k, v)
			}
		},
		eventGet: func(ctx plugin.HandlerCtx) {},
	}
	//Push仅用于开发阶段
	langAPi = plugin.TRGroup{
		"add": {
			TV: plugin.NV{"[KV] [0x01] -> [0x02]"},
			RV: plugin.NV{"[KV] [0x01] & [0x02] & [add]"},
		},
		"del": {
			TV: plugin.NV{"[KV] [0x01] !-> [0x02]"},
			RV: plugin.NV{"[KV] k & v & [Del]"},
		},
		"get": {
			TV: plugin.NV{
				//[0x01]:dist;[0x02]:key
				//the result is in [0x01]
				"check what is [0x02]",
				"[KV] [0x01] & [0x02] & [HGet]",
			},
			RV: plugin.NV{
				"[0x02] is [0x01]",
			},
		},
	}
	engine *Engine
)

func (kv *KVNode) Execute() error {
	return plugin.Execute(head, kv.Parse, &plugin.HandlerCtxBase{}, paramHandlers)
}
func (kv *KVNode) Parse() (m.H, error) {
	return plugin.ParseFunc(kv.Base, paramConfig, paramKeys)
}

// Extract [KV] dist & key & [HGet]
// [KV] dist & value & [Push]
// the result will be put in dist
// [HGet]:need get(key)
// [Push]: push value into dist directly
func (kv *KVNode) Extract() (variable.Item, error) {
	head := "Extract: "

	params, err := kv.Parse()
	if err != nil {
		return nil, fmt.Errorf("%s %s", head, err.Error())
	}

	event, ok := params[plugin.Event].(string)
	if ok {
		var (
			varb    variable.Item
			errMsg  string
			dist, k = engine.KV(params)
		)

		switch event {
		case eventGet:
			v := engine.barn.Get(k)
			if v != "" {
				varb = variable.New(dist, k)
			} else {
				errMsg = log.NotFound(k).Error()
			}

		case eventPush:
			varb = variable.New(dist, k)

		default:
			errMsg = log.NotExist(plugin.Event, event).Error()
		}

		if errMsg != "" {
			return varb, fmt.Errorf("%s %s", head, errMsg)
		}

		return varb, nil
	}

	return nil, log.NotFound(head + plugin.Event)
}
func nodeCreator(b node.Base) node.Item {
	return &KVNode{
		Base: b,
	}
}
func New(e *Engine) plugin.Item {
	if e != nil {
		engine = e
	}
	return &plugin.Base{
		ParamCfg:    paramConfig,
		NodeCreator: nodeCreator,
		LangAPI:     langAPi,
		Name:        head,
	}
}
