package checker

import (
	"fmt"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/plugin"
)

type Context struct {
	Event   string
	Payload []string
}
type Handler func(*Context) bool
type Registry interface {
	Register(event string, method Handler)
	Method(event string) (Handler, bool)
}
type registry struct {
	items map[string]Handler
}

func (r *registry) Method(event string) (Handler, bool) {
	method, ok := r.items[event]
	return method, ok
}

func (c *Context) Set(params map[string]any) {
	if params == nil {
		return
	}
	if event, ok := params[plugin.Event]; ok {
		c.Event = event.(string)
	}
	if payload, ok := params[ParamPayload]; ok {
		c.Payload = payload.([]string)
	}
}

// Register [Handler] handlerName & Parameter_1 & Parameter_2 & Parameter_3...
func (r *registry) Register(event string, method Handler) {
	if event == "" {
		return
	}
	if method == nil {
		return
	}
	r.items[event] = method
}

var (
	Head        = "[BetterCheck]"
	ParamConfig = &plugin.ParamCfg{
		Head:        Head,
		ParamLength: -1,
	}
	ParamPayload = "[payload]"
	ParamKeys    = map[string]int{
		plugin.Event: 0,
		ParamPayload: -1,
	}
)

type BetterCheckNode struct {
	node.Base
	registry Registry
}

func (bc *BetterCheckNode) check() error {
	ctx := &Context{}
	params, err := bc.Parse()
	if err != nil {
		return fmt.Errorf("%s %s", Head, err.Error())
	}
	ctx.Set(params)
	if check, ok := bc.registry.Method(ctx.Event); ok {
		bc.SetState(check(ctx))
	}
	return nil
}
func (bc *BetterCheckNode) Parse() (m.H, error) {
	return plugin.ParseFunc(bc.Base, ParamConfig, ParamKeys)
}
func NodeFactory(e Registry) node.Creator {
	return func(attr node.Base) node.Item {
		return &BetterCheckNode{
			Base:     attr,
			registry: e,
		}
	}
}
func NewChecker(e Registry) plugin.Item {
	return &plugin.Base{
		NodeCreator: NodeFactory(e),
		Name:        Head,
		ParamCfg:    ParamConfig,
		LangAPI:     make(plugin.TRGroup),
	}
}
func NewRegistry() Registry {
	return &registry{
		items: make(map[string]Handler),
	}
}
