package positioner

import (
	"errors"
	"goldenglow/m"
	"goldenglow/pkg/registry"
	"goldenglow/variable"
	"testing"
)

// mockRepository implements storage.Repository for testing
type mockRepository struct {
	data map[string]m.Hash
}

func (mr *mockRepository) HGet(tag string) (m.Hash, error) {
	if hash, ok := mr.data[tag]; ok {
		return hash, nil
	}
	return nil, errors.New("not found")
}

func (mr *mockRepository) HSet(tag string, value m.Hash) error {
	mr.data[tag] = value
	return nil
}

func (mr *mockRepository) Get(key string) (string, error) {
	return "", nil
}

func (mr *mockRepository) Set(key, value string) error {
	return nil
}

func (mr *mockRepository) Shutdown() error {
	return nil
}

func (mr *mockRepository) Init() error {
	return nil
}

// mockNode implements node.Interface for testing
type mockNode struct {
	value string
}

func (mn *mockNode) Execute(state string)                 {}
func (mn *mockNode) Value() string                        { return mn.value }
func (mn *mockNode) VarKeys() []string                    { return nil }
func (mn *mockNode) ToTextWithNoVars(state string) string { return mn.value }
func (mn *mockNode) VarSetRegistry() registry.Interface[variable.Set] {
	return registry.New[variable.Set]()
}

func TestNew(t *testing.T) {
	repo := &mockRepository{data: make(map[string]m.Hash)}
	p := New(repo)
	if p == nil {
		t.Error("expected New to return non-nil positioner")
	}
}

func TestNew_NilRepo(t *testing.T) {
	p := New(nil)
	if p != nil {
		t.Error("expected New with nil repo to return nil")
	}
}

func TestPositioner_ContainerOf(t *testing.T) {
	repo := &mockRepository{data: make(map[string]m.Hash)}
	n := &mockNode{value: "test_node"}

	// Set up test data
	hash := make(m.Hash)
	hash["container1"] = struct{}{}
	hash["container2"] = struct{}{}
	repo.data[prefixT2C+"test_node"] = hash

	p := New(repo)
	result := p.ContainerOf(n)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result) != 2 {
		t.Errorf("expected 2 containers, got %d", len(result))
	}
	if _, ok := result["container1"]; !ok {
		t.Error("expected container1 to be present")
	}
	if _, ok := result["container2"]; !ok {
		t.Error("expected container2 to be present")
	}
}

func TestPositioner_ContainerOf_NoData(t *testing.T) {
	repo := &mockRepository{data: make(map[string]m.Hash)}
	n := &mockNode{value: "nonexistent_node"}

	p := New(repo)
	result := p.ContainerOf(n)

	if result != nil {
		t.Error("expected nil result for nonexistent node")
	}
}

func TestDefault(t *testing.T) {
	p := Default()
	if p == nil {
		t.Error("expected Default to return non-nil positioner")
	}
}
