package plugins

//TODO del the trash:Runner;use real parameters instead of map[string]any
import (
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/plugin"
)

type Context struct {
	data m.H
}

func (c *Context) Get() m.H {
	return c.data
}
func (c *Context) Set(name string, args any) {
	if name != "" {
		c.data[name] = args
	}
}

type Opt func(*Context)
type Caller func(*Context)
type Hook func()
type Registry interface {
	Register(callback Caller)
	RegisterHook(hook Hook)
	Init(options ...Opt)
	ShutDown()
}
type registry struct {
	Registries    []Caller
	ShutdownHooks []Hook
}

func DefaultCTX() *Context {
	return &Context{}
}
func (c *registry) Init(options ...Opt) {
	ctx := DefaultCTX()
	for _, opt := range options {
		opt(ctx)
	}
	for _, caller := range c.Registries {
		if caller != nil {
			caller(ctx)
		}
	}
}
func (c *registry) ShutDown() {
	for _, hook := range c.ShutdownHooks {
		if hook != nil {
			hook()
		}
	}
}
func (c *registry) Register(registry Caller) {
	if registry == nil {
		return
	}
	c.Registries = append(c.Registries, registry)
}
func (c *registry) RegisterHook(hook Hook) {
	if hook == nil {
		return
	}
	c.ShutdownHooks = append(c.ShutdownHooks, hook)
}
func WithLangRegistry(core plugin.LangRegistry) Opt {
	return func(ctx *Context) {
		ctx.Set(plugin.LangRegistryName, core)
	}
}
func WithNodeRegistry(registry node.Registry) Opt {
	return func(ctx *Context) {
		ctx.Set(plugin.NodeRegistry, registry)
	}
}
func NewRegistry() Registry {
	return &registry{
		Registries:    make([]Caller, 0),
		ShutdownHooks: make([]Hook, 0),
	}
}
