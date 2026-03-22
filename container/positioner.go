package container

import (
	"fmt"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/pkg/log"
)

type positioner struct {
	db      Repository
	encoder node.Encoder
}

func (p *positioner) ContainerOf(node node.Item) (m.Hash, error) {
	var (
		tag = prefixT2C + p.encoder.Do(node.Value())
	)
	hashValue, err := p.db.HGet(tag)
	if err != nil {
		return nil, fmt.Errorf("positioner: %w", log.NotFound(tag))
	}
	return hashValue, nil
}
func NewPositioner(db Repository, encoder node.Encoder) (Positioner, error) {
	head := "positioner init"
	if encoder == nil {
		return nil, fmt.Errorf("%s: encoder is nil", head)
	}
	if db == nil {
		return nil, fmt.Errorf("%s: db is nil", head)
	}
	return &positioner{
		db:      db,
		encoder: encoder,
	}, nil
}
