package fetcher

import (
	"goldenglow/m"
	"goldenglow/pkg/node"
	"goldenglow/storage"
	"goldenglow/utils"
)

var (
	prefixC2T = "C->T:" // Container -> T Node
	prefixC2R = "C->R:" // Container -> R Node
)

type Interface interface {
	T(hash string) m.Map[node.Interface]
	R(hash string) m.Map[node.Interface]
}
type fetcher struct {
	db       storage.Repository
	nFactory node.Factory
}

func (f *fetcher) T(hashValue string) m.Map[node.Interface] {
	var (
		tag = prefixC2T + hashValue
	)
	return f.toNode(tag)
}

func (f *fetcher) R(hashValue string) m.Map[node.Interface] {
	var (
		tag = prefixC2R + hashValue
	)
	return f.toNode(tag)
}

func (f *fetcher) toNode(tag string) m.Map[node.Interface] {
	nvMap, err := f.db.HGet(tag)
	if err != nil {
		return nil
	}
	nodes := make(m.Map[node.Interface], len(nvMap))
	for t := range nvMap {
		nodes[t] = f.nFactory.Create(t)
	}
	return nodes
}
func New(db storage.Repository, f node.Factory) Interface {
	if err := utils.NotNull("repo", db, "node_factory", f); err != nil {
		return nil
	}
	return &fetcher{
		db:       db,
		nFactory: f,
	}
}
func Default() Interface {
	return New(
		storage.DefaultJSONRepo(),
		node.DefaultFactory,
	)
}
