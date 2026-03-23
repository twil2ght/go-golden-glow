package executor

import "goldenglow/node"

type Parameters map[string]string
type Handler func(params Parameters) error

type Registry interface {
	RunAll() (string, node.Creator)
	Register(name string, method Handler) error
}
