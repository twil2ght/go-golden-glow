package container

import (
	"fmt"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/pkg/log"
	"goldenglow/utils"
)

type positioner struct {
	db      Repository
	encoder node.Encoder
}

func (p *positioner) Encoder() node.Encoder {
	return p.encoder
}

func (p *positioner) ContainerOf(node interface {
	Value() string
}) (m.Hash, error) {
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
	if err := utils.NotNull("repo", db, "encoder", encoder); err != nil {
		return nil, fmt.Errorf("%s: %w", head, err)
	}
	return &positioner{
		db:      db,
		encoder: encoder,
	}, nil
}
