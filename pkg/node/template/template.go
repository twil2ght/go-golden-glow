package template

import (
	"goldenglow/m"
	"goldenglow/pkg/container/positioner"
	"goldenglow/pkg/node"
	"goldenglow/pkg/repo"
	"goldenglow/storage"
)

var (
	banned = "$1 is $2"
)

type Set m.Map[node.Interface]
type Interface interface {
	GetTemplate(n node.Interface, state string) Set
	// BanFilter for testing only
	BanFilter()
}
type Positioner interface {
	ContainerOf(node.Interface) m.Hash
}
type template struct {
	templates  Set
	repo       storage.Repository
	factory    node.Factory
	positioner Positioner
	banFilter  bool
}

func (t *template) BanFilter() {
	t.banFilter = true
}

func (t *template) initTemplate() {
	nodeHash, _ := t.repo.HGet(repo.KeyNodeSet)

	nodeHash = t.filter(nodeHash)
	for nodeValue := range nodeHash {
		n := t.factory.Create(nodeValue)
		if n != nil {
			t.templates[nodeValue] = n
		}
	}
}
func (t *template) filter(nodeHash m.Hash) m.Hash {
	if t.banFilter {
		return nodeHash
	}
	for nv := range nodeHash {
		n := t.factory.Create(nv)
		if cHash := t.positioner.ContainerOf(n); len(cHash) == 0 {
			if _, ok := n.(*node.Node); ok {
				delete(nodeHash, nv)
			}
		}
	}
	return nodeHash
}
func (t *template) GetTemplate(n node.Interface, state string) Set {
	matches := make(Set)
	if !AllowedToGetTemplate(n) {
		return matches
	}
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
func AllowedToGetTemplate(n node.Interface) bool {
	if _, ok := n.(*node.Node); ok {
		return true
	}
	return false
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
		node.DefaultFactory,
		positioner.Default(),
	)
}
