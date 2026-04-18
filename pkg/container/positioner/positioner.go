package positioner

import (
	"goldenglow/m"
	"goldenglow/pkg/node"
	"goldenglow/storage"
	"goldenglow/utils"
)

var (
	prefixT2C = "T->C:" // T Node -> Container
)

type Interface interface {
	ContainerOf(node.Interface) m.Hash
}
type positioner struct {
	repo storage.Repository
}

func (p *positioner) ContainerOf(n node.Interface) m.Hash {
	var (
		tag = prefixT2C + n.Value()
	)
	hashValue, err := p.repo.HGet(tag)
	if err != nil {
		return nil
	}
	return hashValue
}
func New(repo storage.Repository) Interface {
	if err := utils.NotNull("repo", repo); err != nil {
		return nil
	}
	return &positioner{
		repo: repo,
	}
}
func Default() Interface {
	return New(storage.DefaultJSONRepo())
}
