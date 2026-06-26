package template

import (
	"goldenglow/m"
	"goldenglow/pkg/brainsaver"
	"goldenglow/pkg/container/positioner"
	"goldenglow/pkg/database"
	"goldenglow/pkg/node"
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
	repo       database.Repository
	factory    node.Factory
	positioner Positioner
	banFilter  bool
}

func (t *template) BanFilter() {
	t.banFilter = true
}

func (t *template) initTemplate() {
	nodeHash, _ := t.repo.HGet(brainsaver.KeyNodeSet)

	nodeHash = t.PreFilter(nodeHash)
	for nodeValue := range nodeHash {
		n := t.factory.Create(nodeValue)
		if n != nil {
			t.templates[nodeValue] = n
		}
	}
}
func (t *template) PreFilter(nodeHash m.Hash) m.Hash {
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
	t.initTemplate()
	t.templates = MidFilter(n, t.templates)
	raw := n.ToTextWithNoVars(state)
	for key, e := range t.templates {
		if ok, vars := MatchTemplate(raw, e.Value()); ok {
			e.VarSetRegistry().Register(raw, vars)
			matches[key] = e
			//if n.Value() == "[repo] [GET] [Grammar:Property:Word:Noun] $2 @Caller $3 [Verify:Object]" {
			//	log.Default().Debug("match template", "node", n.Value(), "template", e.Value())
			//}
		}
	}
	return PostFilter(matches)
}
func MidFilter(n node.Interface, set Set) Set {
	var isDefault bool
	_, isDefault = n.(*node.Node)
	for nv := range set {
		_, typeOK := set[nv].(*node.Node)
		if typeOK != isDefault {
			delete(set, nv)
		}
	}
	return set
}
func PostFilter(set Set) Set {
	for key := range set {
		//log.Default().Debug("checking conflict rule", "node", key)
		conflictRule, _ := DefaultConflictManager.Get(key)
		if conflictRule != nil {
			//log.Default().Debug("applying conflict rule", "node", key)
			conflictRule(set[key], set)
		}
	}
	//for key := range set {
	//	log.Default().Debug("remaining node", "node", key)
	//}
	return set
}
func New(repo database.Repository, factory node.Factory, positioner Positioner) Interface {
	return &template{
		templates:  Set{},
		factory:    factory,
		positioner: positioner,
		repo:       repo,
	}
}
func Default() Interface {
	return New(
		database.DefaultJSONRepo(),
		node.DefaultFactory,
		positioner.Default(),
	)
}
