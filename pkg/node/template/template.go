package template

import (
	"goldenglow/container"
	"goldenglow/m"
	"goldenglow/pkg/node"
	"goldenglow/storage"
)

var (
	banned = "$1 is $2"
)

type Set m.Map[node.Interface]
type Interface interface {
	GetTemplate(n node.Interface, state string) Set
}
type Positioner interface {
	GetContainerHashByTrigger(node.Interface) m.Hash
}
type template struct {
	templates  Set
	repo       storage.Repository
	factory    node.Factory
	positioner Positioner
}

func (t *template) initTemplate() {
	nodeHash, _ := t.repo.HGet(container.KeyNodeSet)

	nodeHash = t.filter(nodeHash)
	for nodeValue := range nodeHash {
		n := t.factory.Create(nodeValue)
		if n != nil {
			t.templates[nodeValue] = n
		}
	}
}
func (t *template) filter(nodeHash m.Hash) m.Hash {
	for nv := range nodeHash {
		n := t.factory.Create(nv)
		if cHash := t.positioner.GetContainerHashByTrigger(n); len(cHash) == 0 {
			delete(nodeHash, nv)
		}
	}
	return nodeHash
}
func (t *template) GetTemplate(n node.Interface, state string) Set {
	matches := make(Set)
	t.initTemplate()
	raw := n.ToTextWithNoVars(state)
	for key, e := range t.templates {
		if ok, vars := matchTemplate(raw, e.Value()); ok {
			e.VarSetRegistry().Register(raw, vars)
			matches[key] = e
		}
	}
	return t.FilterBanned(matches)
}
func (t *template) FilterBanned(set Set) Set {
	if _, ok := set[banned]; ok && len(set) > 1 {
		delete(set, banned)
	}
	return set
}
func New(repo storage.Repository, factory node.Factory, positioner Positioner) Interface {
	return &template{
		templates:  Set{},
		factory:    factory,
		positioner: positioner,
		repo:       repo,
	}
}
func Default() Interface {
	return New(
		storage.DefaultJSONRepo(),
		node.DefaultFactory(),
		nil,
	)
}
