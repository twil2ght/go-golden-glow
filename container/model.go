package container

import (
	"errors"
	"goldenglow/m"
	"goldenglow/node"
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
}
type factory struct {
	encoder    node.Encoder
	fetcher    Fetcher
	positioner Positioner
}

func (f *factory) New(hashValue string) (Item, error) {
	return New(hashValue, f.fetcher, f.encoder)
}
func NewFactory(fetcher Fetcher, encoder node.Encoder, p Positioner) (Factory, error) {
	if fetcher == nil {
		return nil, errors.New("fetcher is nil")
	}
	if encoder == nil {
		return nil, errors.New("encoder is nil")
	}
	if p == nil {
		return nil, errors.New("positioner is nil")
	}
	return &factory{
		fetcher:    fetcher,
		encoder:    encoder,
		positioner: p,
	}, nil
}
func (f *factory) Encoder() node.Encoder  { return f.encoder }
func (f *factory) Positioner() Positioner { return f.positioner }
