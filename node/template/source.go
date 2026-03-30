package template

import (
	"goldenglow/container"
	"goldenglow/m"
	"goldenglow/node"
	"goldenglow/storage"
)

type source struct {
	repo    storage.Repository
	factory node.Factory
}

func (s *source) GetTemplates() (node.Set, error) {
	nodeHash, err := s.repo.HGet(container.KeyNodeSet)
	if err != nil {
		return nil, err
	}
	nodeHash = s.filter(nodeHash)
	nodes := make(node.Set)
	for nodeValue := range nodeHash {
		n, err := s.factory.NewFromPool(nodeValue)
		if err != nil {
			return nil, err
		}
		nodes[nodeValue] = n
	}
	return nodes, nil
}
func (s *source) filter(nodeValue m.Hash) m.Hash {
	return nodeValue
}
