package variable

import (
	"goldenglow/m"
	"goldenglow/pkg/log"
	"regexp"
)

type Item interface {
	Name() string
	Value() string
	Set(value string) error
	OK() bool
	Copy() Item
}

// ValueMap is an interface for items that carry a VariableSetHub
// This allows extractors to return multiple variable sets that can be
// inherited by result nodes
type ValueMap interface {
	Item
	ValueMap() m.Hash
}
type Parser func(strWithVariables string, variables Set, strict bool) (string, error)
type Set map[string]Item
type Base struct {
	name  string
	value string
}

var (
	VarReg = regexp.MustCompile(`\$\d+`)
)

func (b *Base) Name() string {
	return b.name
}
func (b *Base) Value() string {
	return b.value
}
func (b *Base) OK() bool {
	return b.value != ""
}

func (b *Base) Set(value string) error {
	if value != "" {
		b.value = value
		return nil
	}
	return log.EmptyStrErr()
}

func (b *Base) Copy() Item {
	return New(b.Name(), b.Value())
}

func New(k, v string) Item {
	if k == "" {
		k = "you get an empty key"
	}
	return &Base{
		name:  k,
		value: v,
	}
}

// valueMapItem is a variable item that carries a state hub (map[string]Set)
// This allows extractors to return multiple variable sets for result nodes to inherit
type valueMapItem struct {
	Base
	hash m.Hash
}

func (s *valueMapItem) ValueMap() m.Hash {
	return s.hash
}

// NewValueMap creates a new variable item with a state hub
func NewValueMap(k, v string, values m.Hash) ValueMap {
	if k == "" {
		k = "you get an empty key"
	}
	return &valueMapItem{
		Base: Base{
			name:  k,
			value: v,
		},
		hash: values,
	}
}
func Copy(target Set) Set {
	dist := make(Set, len(target))
	for k, v := range target {
		dist[k] = New(k, v.Value())
	}
	return dist
}

// HasCycle checks if the variable set contains any circular references
// Returns true if a cycle exists (e.g., $1 -> $2 -> $1)
func (s Set) HasCycle() bool {
	// Track visited nodes and nodes in current recursion stack
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(varName string) bool
	dfs = func(varName string) bool {
		visited[varName] = true
		recStack[varName] = true

		if item, ok := s[varName]; ok {
			// Find all variables referenced in this variable's value
			refs := VarReg.FindAllString(item.Value(), -1)
			for _, ref := range refs {
				if !visited[ref] {
					if dfs(ref) {
						return true
					}
				} else if recStack[ref] {
					// Found a back edge - cycle detected
					return true
				}
			}
		}

		recStack[varName] = false
		return false
	}

	// Check all variables in the set
	for varName := range s {
		if !visited[varName] {
			if dfs(varName) {
				return true
			}
		}
	}
	return false
}

func ToRawText(target string, variables Set, strict bool) (string, error) {
	var (
		changed  = true
		res      = target
		prev     = target
		shutdown = false
		source   = ""
	)
	if variables.HasCycle() {
		return target, log.NewErr("variable has cycle")
	}
	for changed {
		changed = false
		prev = res
		res = VarReg.ReplaceAllStringFunc(res, func(s string) string {
			if varb, ok := variables[s]; ok {
				return varb.Value()
			}
			if strict {
				shutdown = true
				source = s
			}
			return s
		})
		if shutdown {
			return "", log.NotExist("variable", source)
		}
		if res != prev {
			changed = true
		}
	}
	return res, nil
}
