package node

import (
	"goldenglow/variable"
)

type Creator func(base Base) Item

type Checkable interface {
	Check() error
}
type Extractable interface {
	Extract() (variable.Item, error)
}
type VarReplacer interface {
	ReplaceAllStringFunc(src string, repl func(string) string) string
}
type Encoder interface {
	Do(nodeValue string) string
	Match(nv1, nv2 string) bool
}
type LightRepository interface {
	Get(key string) (string, error)
}
type Registry interface {
	Register(name string, method Creator) error
}
type Pool interface {
	Add(node Item)
}
type Factory interface {
	New(nodeValue string) (Item, error)
	Registry
	Pool
}
