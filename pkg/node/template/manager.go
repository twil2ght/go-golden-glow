package template

import (
	"goldenglow/pkg/node"
	"goldenglow/pkg/registry"
)

type ConflictRule func(origin, tpl node.Interface) bool
type ConflictManager interface {
	Register(valueOfOriginalNode string, rule ConflictRule)
	registry.Interface[ConflictRule]
}
type manager struct {
	registry.Interface[ConflictRule]
}

func (m *manager) Register(valueOfOriginalNode string, rule ConflictRule) {
	m.Interface.Register(valueOfOriginalNode, rule)
}
func NewConflictManager() ConflictManager {
	return &manager{registry.New[ConflictRule]()}
}

var (
	DefaultConflictManager = NewConflictManager()
)
