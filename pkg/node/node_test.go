package node

import (
	"goldenglow/pkg/variable"
	"testing"
)

func TestNew(t *testing.T) {
	node := New("test_value")
	if node == nil {
		t.Fatal("expected New to return non-nil node")
	}
	if node.Value() != "test_value" {
		t.Errorf("expected value 'test_value', got '%s'", node.Value())
	}
}

func TestNode_Value(t *testing.T) {
	node := New("hello world")
	if node.Value() != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", node.Value())
	}
}

func TestNode_Execute(t *testing.T) {
	node := New("test")
	// Execute does nothing, just ensure it doesn't panic
	node.Execute("some_state")
}

func TestNode_VarKeys(t *testing.T) {
	tests := []struct {
		value    string
		expected []string
	}{
		{"no variables", []string{}},
		{"$1", []string{"$1"}},
		{"$1 $2 $3", []string{"$1", "$2", "$3"}},
		{"text $1 more $2", []string{"$1", "$2"}},
		{"$10 $2 $1", []string{"$10", "$2", "$1"}},
	}

	for _, tt := range tests {
		node := New(tt.value)
		keys := node.VarKeys()
		if len(keys) != len(tt.expected) {
			t.Errorf("for value '%s', expected %d keys, got %d: %v", tt.value, len(tt.expected), len(keys), keys)
			continue
		}
		for i, expected := range tt.expected {
			if keys[i] != expected {
				t.Errorf("for value '%s', expected key[%d]='%s', got '%s'", tt.value, i, expected, keys[i])
			}
		}
	}
}

func TestNode_ToTextWithNoVars(t *testing.T) {
	node := New("hello $1 world $2")

	// Register a variable set for "state1"
	varSet := make(variable.Set)
	varSet["$1"] = variable.New("$1", "beautiful")
	varSet["$2"] = variable.New("$2", "universe")
	node.VarSetRegistry().Register("state1", varSet)

	result := node.ToTextWithNoVars("state1")
	expected := "hello beautiful world universe"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestNode_ToTextWithNoVars_NoVars(t *testing.T) {
	node := New("hello world")
	result := node.ToTextWithNoVars("state1")
	if result != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", result)
	}
}

func TestNode_ToTextWithNoVars_MissingState(t *testing.T) {
	node := New("hello $1")
	// No variable set registered for "missing_state"
	result := node.ToTextWithNoVars("missing_state")
	// Should return the original string since varSet is nil/empty
	if result != "hello $1" {
		t.Errorf("expected 'hello $1', got '%s'", result)
	}
}

func TestNode_ToTextWithNoVars_PartialVars(t *testing.T) {
	node := New("hello $1 $2")

	varSet := make(variable.Set)
	varSet["$1"] = variable.New("$1", "world")
	// $2 not defined
	node.VarSetRegistry().Register("state1", varSet)

	result := node.ToTextWithNoVars("state1")
	// Since strict=false, undefined vars should remain
	if result != "hello world $2" {
		t.Errorf("expected 'hello world $2', got '%s'", result)
	}
}

func TestNode_VarSetRegistry(t *testing.T) {
	node := New("test")
	registry := node.VarSetRegistry()
	if registry == nil {
		t.Error("expected VarSetRegistry to return non-nil registry")
	}
	if registry.Len() != 0 {
		t.Errorf("expected empty registry, got len=%d", registry.Len())
	}
}
