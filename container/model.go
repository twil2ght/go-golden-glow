package container

import (
	"goldenglow/m"
	"goldenglow/node"
	"regexp"
)

const (
	prefixC2T = "C->T:" // Container -> T Node
	prefixC2R = "C->R:" // Container -> R Node
	prefixT2C = "T->C:" // T Node -> Container
	prefixR2C = "R->C:" // R Node -> Container
	TypeT     = "T"
	TypeR     = "R"
)

type Repository interface {
	HGet(tag string) (m.Hash, error)
	HSet(tag string, value m.Hash) error
}
type Positioner interface {
	ContainerOf(external node.Item) (m.Hash, error)
	Encoder() node.Encoder
}

type Fetcher interface {
	TNode(hashValue string) (node.Set, error)
	RNode(hashValue string) (node.Set, error)
}

type Store interface {
	Save(tv, rv m.Hash) error
}

type Factory interface {
	New(hashValue string) (Item, error)
	Encoder() node.Encoder
	Positioner() Positioner
	WithVarReg(variableReg *regexp.Regexp)
	ResetNodePool()
}
