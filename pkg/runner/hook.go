package runner

import (
	"goldenglow/pkg/registry"
	"goldenglow/pkg/tracer"
)

type IdleHandler func() bool
type Hook interface {
	OnRegisterIdleHandler(mgr registry.Interface[IdleHandler])
}

type TraceHandler func(event tracer.Event)
type TraceHook interface {
	OnRegisterTraceHandler(mgr registry.Interface[TraceHandler])
}

var DefaultManager = registry.New[IdleHandler]()
var DefaultTraceManager = registry.New[TraceHandler]()
