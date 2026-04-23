package positioner

import (
	"goldenglow/m"
	"goldenglow/pkg/database"
	"goldenglow/pkg/node"
	"goldenglow/utils"
)

var (
	prefixT2C = "T->C:" // T Node -> Container
)

type Interface interface {
	ContainerOf(node.Interface) m.Hash
}
type positioner struct {
	repo database.Repository
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
func New(repo database.Repository) Interface {
	if err := utils.NotNull("repo", repo); err != nil {
		return nil
	}
	return &positioner{
		repo: repo,
	}
}
func Default() Interface {
	return New(database.DefaultJSONRepo())
}
