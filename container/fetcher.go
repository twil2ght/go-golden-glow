package container

import (
	"fmt"
	"goldenglow/node"
	"goldenglow/utils"
)

type fetcher struct {
	db       Repository
	nFactory node.Factory
}

func (f *fetcher) TNode(hashValue string) (node.Set, error) {
	var (
		tag = prefixC2T + hashValue
	)
	return f.toNode(tag)
}

func (f *fetcher) RNode(hashValue string) (node.Set, error) {
	var (
		tag = prefixC2R + hashValue
	)
	return f.toNode(tag)
}

func (f *fetcher) toNode(tag string) (node.Set, error) {
	head := "Fetcher"
	nvMap, err := f.db.HGet(tag)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", head, err)
	}

	nodes := make(node.Set, len(nvMap))
	for t := range nvMap {
		n, err := f.nFactory.New(t)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", t, err)
		}
		nodes[t] = n
	}

	return nodes, nil
}
func NewFetcher(db Repository, f node.Factory) (Fetcher, error) {
	head := "Fetcher init"
	if err := utils.NotNull("repo", db, "node_factory", f); err != nil {
		return nil, fmt.Errorf("%s: %w", head, err)
	}
	return &fetcher{
		db:       db,
		nFactory: f,
	}, nil
}
