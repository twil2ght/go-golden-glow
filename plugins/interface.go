package plugins

import (
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/pinkcat"
	"goldenglow/plugin"
	"sync"
)

type Runner interface {
	Run() error
	RunningState() bool
	GetLock() *sync.Mutex
	GetUserID() int
	SetDecorator(pinkcat.Decorator)
	SetValidator(pinkcat.Decorator)
	SetRunningState(bool)
	GetDecorator() pinkcat.Decorator
	GetValidator() pinkcat.Decorator
}
type Context struct {
	Runner Runner
	UserID int
	data   m.H
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
	return &Context{
		UserID: 0,
	}
}
func (c *registry) Init(options ...Opt) {
	ctx := DefaultCTX()
	for _, opt := range options {
		opt(ctx)
	}
	if ctx.Runner == nil {
		panic("invalid runner")
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
func WithRunner(r Runner) Opt {
	return func(ctx *Context) {
		ctx.Runner = r
	}
}
func WithUserID(userID int) Opt {
	return func(ctx *Context) {
		ctx.UserID = userID
	}
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
