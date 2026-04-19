package container

import (
	"goldenglow/m"
	"goldenglow/pkg/container/fetcher"
	"goldenglow/pkg/node"
	"goldenglow/utils"
	"goldenglow/variable"
)

type Checkable interface {
	Check(state string) bool
}
type Extractable interface {
	Extract(state string) variable.ValueMap
	KeyDist() string
}
type Interface interface {
	T() m.Map[node.Interface]
	R() (nodeMap m.Map[node.Interface], stateMap m.Map[[]string])
	Forward(t node.Interface, state string) bool
}
type container struct {
	hash                string
	t                   m.Map[node.Interface]
	r                   m.Map[node.Interface]
	s                   m.Map[[]string]
	fetcher             fetcher.Interface
	varSet              variable.Set
	extractDist         string
	extractDistValueMap variable.ValueMap
}

func (c *container) T() m.Map[node.Interface] {
	return c.t
}
func (c *container) R() (m.Map[node.Interface], m.Map[[]string]) {
	return c.r, c.s
}
func (c *container) Forward(t node.Interface, state string) bool {
	if !c.fetch() {
		return false
	}
	varSet, _ := t.VarSetRegistry().Get(state)
	if !c.findCompatibleVarSet(t, varSet) {
		return false
	}
	if !c.handleSpecialT() {
		return false
	}
	c.FinalResults()
	return true
}
func (c *container) fetch() bool {
	c.t = c.fetcher.T(c.hash)
	c.r = c.fetcher.R(c.hash)
	if c.t == nil || c.r == nil {
		return false
	}
	return true
}
func (c *container) findCompatibleVarSet(t node.Interface, varSet variable.Set) bool {
	if varSet != nil {
		c.varSet = varSet
	}
	return mergeVariables(t, c.t, c.varSet)
}
func (c *container) handleSpecialT() bool {
	for _, t := range c.t {
		if _, ok := t.(Checkable); ok {
			if !c.handleChecker(t) {
				return false
			}
		}
		if _, ok := t.(Extractable); ok {
			if !c.handlerExtractor(t) {
				return false
			}
		}
	}
	return true
}
func (c *container) handleChecker(t node.Interface) bool {
	checker := t.(Checkable)
	state, varSet := c.getStateFromVarSetOfContainer(t.VarKeys())
	t.VarSetRegistry().Register(state, varSet)
	return checker.Check(state)
}
func (c *container) handlerExtractor(t node.Interface) bool {
	extractor := t.(Extractable)
	c.extractDist = extractor.KeyDist()
	keys := utils.Filter(t.VarKeys(), c.extractDist)
	state, varSet := c.getStateFromVarSetOfContainer(keys)
	t.VarSetRegistry().Register(state, varSet)
	varSetMap := extractor.Extract(state)
	c.extractDistValueMap = varSetMap
	return varSet != nil
}
func (c *container) getStateFromVarSetOfContainer(keys []string) (string, variable.Set) {
	varSet := make(variable.Set)
	for _, key := range keys {
		varSet[key] = c.varSet[key]
	}
	return node.GenVariableState(varSet), varSet
}
func (c *container) getStateFromVarSetOfContainerForExtractDist(keysFiltered []string) m.Map[variable.Set] {
	varSetMap := make(m.Map[variable.Set])
	varSetStart := make(variable.Set)
	for _, key := range keysFiltered {
		varSetStart[key] = c.varSet[key]
	}
	for value := range c.extractDistValueMap.Get() {
		varSetCopy := variable.Copy(varSetStart)
		varSetCopy[c.extractDist] = variable.New(c.extractDist, value)
		varSetMap[node.GenVariableState(varSetCopy)] = varSetCopy
	}
	return varSetMap
}
func (c *container) FinalResults() {
	for _, r := range c.r {
		varSetRegistry := r.VarSetRegistry()
		stateSlice := c.s[r.Value()]

		if c.extractDist != "" && utils.Contain(r.VarKeys(), c.extractDist) {
			varSetMap := c.getStateFromVarSetOfContainerForExtractDist(utils.Filter(r.VarKeys(), c.extractDist))
			for state, varSet := range varSetMap {
				varSetRegistry.Register(state, varSet)
				stateSlice = append(stateSlice, state)
			}
		} else {
			state, varSet := c.getStateFromVarSetOfContainer(r.VarKeys())
			varSetRegistry.Register(state, varSet)
			stateSlice = []string{state}
		}

		c.s[r.Value()] = stateSlice
	}
}

// New creates a new container
func New(hash string, p fetcher.Interface) Interface {
	return &container{
		hash:    hash,
		fetcher: p,
		s:       make(m.Map[[]string]),
		t:       make(m.Map[node.Interface]),
		r:       make(m.Map[node.Interface]),
		varSet:  make(variable.Set),
	}
}

// NewWithDefaultFetcher creates a new container with the default fetcher
func NewWithDefaultFetcher(hash string) Interface {
	return New(hash, fetcher.Default())
}
