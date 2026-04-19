package variable

import (
	"goldenglow/m"
	"goldenglow/pkg/log"
	"regexp"
	"sync"
)

// Item is an interface for variable
type Item interface {
	// Name return the name of the variable
	Name() string
	// Value return the value of the variable
	Value() string
	// Set the value of the variable
	Set(value string) error
	// OK return true if the variable is valid(value !="")
	OK() bool
	// Copy return a copy of the variable item
	Copy() Item
}

// ValueMap is an interface for items that carry a VariableSetHub
// This allows extractors to return multiple variable sets that can be
// inherited by result nodes
type ValueMap interface {
	Get() m.Hash
}

// Parser parse a given string with variables to a one without variables based on the given variables
// strict: if true, return error if variable not found
type Parser func(strWithVariables string, variables Set, strict bool) (string, error)

// Set is a map of variable
type Set map[string]Item

// Base is a base struct for variable
type Base struct {
	name  string
	value string
	mu    *sync.RWMutex
}

var (
	//VarReg is a regular expression for variable(e.g. $1 $2...)
	VarReg = regexp.MustCompile(`\$\d+`)
)

func (b *Base) Name() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.name
}
func (b *Base) Value() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.value
}
func (b *Base) OK() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.value != ""
}

func (b *Base) Set(value string) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if value != "" {
		b.value = value
		return nil
	}
	return log.EmptyStrErr()
}

func (b *Base) Copy() Item {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return New(b.Name(), b.Value())
}

// New create a new variable with the given key and value
func New(k, v string) Item {
	if k == "" {
		k = "you get an empty key"
	}
	return &Base{
		name:  k,
		value: v,
		mu:    &sync.RWMutex{},
	}
}

// valueMapItem implements ValueMap interface
type valueMapItem struct {
	Item
	hash m.Hash
}

func (s *valueMapItem) Get() m.Hash {
	return s.hash
}

// NewValueMap creates a new variable item with a state hub
func NewValueMap(k, v string, values m.Hash) ValueMap {
	if k == "" {
		k = "you get an empty key"
	}
	return &valueMapItem{
		Item: New(k, v),
		hash: values,
	}
}

// Copy create a copy of the variable set
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

// ToRawText implementing Parser
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
			if e, ok := variables[s]; ok {
				return e.Value()
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
