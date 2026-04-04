package container

import (
	"goldenglow/node"
	"goldenglow/variable"
	"testing"
)

// mockNode implements node.Item for testing
type mockNode struct {
	value            string
	state            bool
	variables        variable.Set
	variableSetHub   map[string]variable.Set
	variableStateMap map[string]map[string]bool
}

func (m *mockNode) Value() string                                { return m.value }
func (m *mockNode) Variables() variable.Set                      { return variable.Copy(m.variables) }
func (m *mockNode) VariableKeys() []string                       { return nil }
func (m *mockNode) SetState(state bool)                          { m.state = state }
func (m *mockNode) SetVariable(vars variable.Set) error          { m.variables = vars; return nil }
func (m *mockNode) OK() bool                                     { return m.state }
func (m *mockNode) Execute() error                               { return nil }
func (m *mockNode) VariableStateMap() map[string]map[string]bool { return m.variableStateMap }
func (m *mockNode) VariableSetFromHub(state string) variable.Set { return m.variableSetHub[state] }
func (m *mockNode) VariableStateExecute() map[string]bool        { return nil }
func (m *mockNode) MarkExecuteState(state string)                {}
func (m *mockNode) VariableSetHub() map[string]variable.Set      { return m.variableSetHub }
func (m *mockNode) MarkDone(state, cHash string)                 {}
func (m *mockNode) ToText() (string, error)                      { return m.value, nil }
func (m *mockNode) Reset()                                       {}

func TestFindCompatibleVariableSet(t *testing.T) {
	// Test with empty base variables
	t.Run("empty base variables", func(t *testing.T) {
		base := &Base{variables: make(variable.Set)}

		nodes := []node.Item{
			&mockNode{
				value: "if A is in B",
				variableSetHub: map[string]variable.Set{
					"state1": {"$A": variable.New("$A", "January"), "$B": variable.New("$B", "winter")},
					"state2": {"$A": variable.New("$A", "March"), "$B": variable.New("$B", "spring")},
				},
			},
			&mockNode{
				value: "if B is a season",
				variableSetHub: map[string]variable.Set{
					"state1": {"$B": variable.New("$B", "winter")},
					"state2": {"$B": variable.New("$B", "spring")},
				},
			},
		}

		result := base.findCompatibleVariableSet(nodes)

		// Should return first compatible set (January/winter)
		if result == nil {
			t.Fatal("expected result, got nil")
		}
		if result["$A"].Value() != "January" {
			t.Errorf("expected $A=January, got %s", result["$A"].Value())
		}
		if result["$B"].Value() != "winter" {
			t.Errorf("expected $B=winter, got %s", result["$B"].Value())
		}
	})

	// Test with base variables set (simulating trigger with $B=winter)
	t.Run("with base variable $B=winter", func(t *testing.T) {
		base := &Base{
			variables: variable.Set{"$B": variable.New("$B", "winter")},
		}

		nodes := []node.Item{
			&mockNode{
				value: "if A is in B",
				variableSetHub: map[string]variable.Set{
					"state1": {"$A": variable.New("$A", "January"), "$B": variable.New("$B", "winter")},
					"state2": {"$A": variable.New("$A", "March"), "$B": variable.New("$B", "spring")},
				},
			},
		}

		result := base.findCompatibleVariableSet(nodes)

		// Should only return January/winter since base has $B=winter
		if result == nil {
			t.Fatal("expected result, got nil")
		}
		if result["$A"].Value() != "January" {
			t.Errorf("expected $A=January, got %s", result["$A"].Value())
		}
		if result["$B"].Value() != "winter" {
			t.Errorf("expected $B=winter, got %s", result["$B"].Value())
		}
	})

	// Test with base variables set (simulating trigger with $B=spring)
	t.Run("with base variable $B=spring", func(t *testing.T) {
		base := &Base{
			variables: variable.Set{"$B": variable.New("$B", "spring")},
		}

		nodes := []node.Item{
			&mockNode{
				value: "if A is in B",
				variableSetHub: map[string]variable.Set{
					"state1": {"$A": variable.New("$A", "January"), "$B": variable.New("$B", "winter")},
					"state2": {"$A": variable.New("$A", "March"), "$B": variable.New("$B", "spring")},
				},
			},
		}

		result := base.findCompatibleVariableSet(nodes)

		// Should only return March/spring since base has $B=spring
		if result == nil {
			t.Fatal("expected result, got nil")
		}
		if result["$A"].Value() != "March" {
			t.Errorf("expected $A=March, got %s", result["$A"].Value())
		}
		if result["$B"].Value() != "spring" {
			t.Errorf("expected $B=spring, got %s", result["$B"].Value())
		}
	})

	// Test empty nodes list
	t.Run("empty nodes list", func(t *testing.T) {
		base := &Base{variables: make(variable.Set)}
		result := base.findCompatibleVariableSet([]node.Item{})
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	// Test when no compatible states exist - should fallback
	t.Run("no compatible states - fallback", func(t *testing.T) {
		base := &Base{
			variables: variable.Set{"$B": variable.New("$B", "summer")}, // No node has $B=summer
		}

		nodes := []node.Item{
			&mockNode{
				value:     "if A is in B",
				variables: variable.Set{"$A": variable.New("$A", "June")},
				variableSetHub: map[string]variable.Set{
					"state1": {"$A": variable.New("$A", "January"), "$B": variable.New("$B", "winter")},
					"state2": {"$A": variable.New("$A", "March"), "$B": variable.New("$B", "spring")},
				},
			},
		}

		result := base.findCompatibleVariableSet(nodes)

		// Fallback should return current variables from the node
		if result == nil {
			t.Fatal("expected fallback result, got nil")
		}
		// The fallback returns n.Variables() which is {$A: "June"}
		if result["$A"].Value() != "June" {
			t.Errorf("expected $A=June from fallback, got %s", result["$A"].Value())
		}
	})
}

func TestIsCompatibleWithBase(t *testing.T) {
	base := &Base{}

	tests := []struct {
		name     string
		vSet     variable.Set
		baseVars variable.Set
		expected bool
	}{
		{
			name:     "compatible - no shared variables",
			vSet:     variable.Set{"$A": variable.New("$A", "January")},
			baseVars: variable.Set{"$B": variable.New("$B", "winter")},
			expected: true,
		},
		{
			name:     "compatible - same value for shared variable",
			vSet:     variable.Set{"$A": variable.New("$A", "January"), "$B": variable.New("$B", "winter")},
			baseVars: variable.Set{"$B": variable.New("$B", "winter")},
			expected: true,
		},
		{
			name:     "incompatible - different value for shared variable",
			vSet:     variable.Set{"$A": variable.New("$A", "January"), "$B": variable.New("$B", "winter")},
			baseVars: variable.Set{"$B": variable.New("$B", "spring")},
			expected: false,
		},
		{
			name:     "empty base vars",
			vSet:     variable.Set{"$A": variable.New("$A", "January")},
			baseVars: variable.Set{},
			expected: true,
		},
		{
			name:     "empty vSet",
			vSet:     variable.Set{},
			baseVars: variable.Set{"$B": variable.New("$B", "winter")},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := base.isCompatibleWithBase(tt.vSet, tt.baseVars)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetCompatibleVariableSets(t *testing.T) {
	base := &Base{
		variables: variable.Set{"$B": variable.New("$B", "winter")},
	}

	node := &mockNode{
		value: "if A is in B",
		variableSetHub: map[string]variable.Set{
			"state1": {"$A": variable.New("$A", "January"), "$B": variable.New("$B", "winter")},
			"state2": {"$A": variable.New("$A", "March"), "$B": variable.New("$B", "spring")},
		},
	}

	result := base.getCompatibleVariableSets(node, base.variables)

	// Should only return state1 since base has $B=winter
	if len(result) != 1 {
		t.Errorf("expected 1 compatible set, got %d", len(result))
	}

	if result[0]["$A"].Value() != "January" {
		t.Errorf("expected $A=January, got %s", result[0]["$A"].Value())
	}
}

func TestMergeTwoStates(t *testing.T) {
	base := &Base{}

	tests := []struct {
		name     string
		setA     variable.Set
		setB     variable.Set
		expected map[string]string
		ok       bool
	}{
		{
			name:     "compatible sets with no overlap",
			setA:     variable.Set{"$A": variable.New("$A", "January")},
			setB:     variable.Set{"$B": variable.New("$B", "winter")},
			expected: map[string]string{"$A": "January", "$B": "winter"},
			ok:       true,
		},
		{
			name:     "compatible sets with same value for shared variable",
			setA:     variable.Set{"$A": variable.New("$A", "January"), "$B": variable.New("$B", "winter")},
			setB:     variable.Set{"$B": variable.New("$B", "winter")},
			expected: map[string]string{"$A": "January", "$B": "winter"},
			ok:       true,
		},
		{
			name:     "incompatible sets with different values for shared variable",
			setA:     variable.Set{"$A": variable.New("$A", "January"), "$B": variable.New("$B", "winter")},
			setB:     variable.Set{"$B": variable.New("$B", "spring")},
			expected: nil,
			ok:       false,
		},
		{
			name:     "empty sets",
			setA:     variable.Set{},
			setB:     variable.Set{},
			expected: map[string]string{},
			ok:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := base.mergeTwoStates(tt.setA, tt.setB)

			if ok != tt.ok {
				t.Errorf("expected ok=%v, got ok=%v", tt.ok, ok)
				return
			}

			if !ok {
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d variables, got %d", len(tt.expected), len(result))
			}

			for key, expectedVal := range tt.expected {
				if item, ok := result[key]; !ok {
					t.Errorf("missing variable %s", key)
				} else if item.Value() != expectedVal {
					t.Errorf("variable %s: expected %s, got %s", key, expectedVal, item.Value())
				}
			}
		})
	}
}

func TestMergeCompatibleStates(t *testing.T) {
	base := &Base{}

	setA := []variable.Set{
		{"$A": variable.New("$A", "January"), "$B": variable.New("$B", "winter")},
		{"$A": variable.New("$A", "March"), "$B": variable.New("$B", "spring")},
	}

	setB := []variable.Set{
		{"$B": variable.New("$B", "winter")},
		{"$B": variable.New("$B", "spring")},
	}

	result := base.mergeCompatibleStates(setA, setB)

	if len(result) != 2 {
		t.Errorf("expected 2 compatible combinations, got %d", len(result))
	}

	// Check that we have both winter and spring combinations
	hasWinter := false
	hasSpring := false
	for _, set := range result {
		if b, ok := set["$B"]; ok {
			if b.Value() == "winter" {
				hasWinter = true
			}
			if b.Value() == "spring" {
				hasSpring = true
			}
		}
	}

	if !hasWinter {
		t.Error("missing winter combination")
	}
	if !hasSpring {
		t.Error("missing spring combination")
	}
}
