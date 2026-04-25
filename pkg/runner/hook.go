package runner

import "goldenglow/pkg/registry"

type IdleHandler func() bool
type Hook interface {
	OnRegisterIdleHandler(mgr registry.Interface[IdleHandler])
}

var DefaultManager = registry.New[IdleHandler]()
