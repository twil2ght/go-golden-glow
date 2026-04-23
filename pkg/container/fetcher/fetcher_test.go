package fetcher

import (
	"errors"
	"goldenglow/m"
	"goldenglow/pkg/node"
	"goldenglow/pkg/registry"
	"goldenglow/pkg/variable"
	"testing"
)

// mockRepository implements database.Repository for testing
type mockRepository struct {
	data map[string]m.Hash
}

func (mr *mockRepository) HGet(tag string) (m.Hash, error) {
	if hash, ok := mr.data[tag]; ok {
		return hash, nil
	}
	return nil, errors.New("not found") // Simulate not found
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

// mockFactory implements node.Factory for testing
type mockFactory struct {
	created map[string]node.Interface
}

func (mf *mockFactory) Create(value string) node.Interface {
	if n, ok := mf.created[value]; ok {
		return n
	}
	// Create a simple mock node
	mockNode := &mockNode{value: value}
	mf.created[value] = mockNode
	return mockNode
}

func (mf *mockFactory) CreatorRegistry() registry.Interface[node.Creator] {
	return registry.New[node.Creator]()
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
	factory := &mockFactory{created: make(map[string]node.Interface)}

	f := New(repo, factory)
	if f == nil {
		t.Error("expected New to return non-nil fetcher")
	}
}

func TestNew_NilRepo(t *testing.T) {
	factory := &mockFactory{created: make(map[string]node.Interface)}
	f := New(nil, factory)
	if f != nil {
		t.Error("expected New with nil repo to return nil")
	}
}

func TestNew_NilFactory(t *testing.T) {
	repo := &mockRepository{data: make(map[string]m.Hash)}
	f := New(repo, nil)
	if f != nil {
		t.Error("expected New with nil factory to return nil")
	}
}

func TestFetcher_T(t *testing.T) {
	repo := &mockRepository{data: make(map[string]m.Hash)}
	factory := &mockFactory{created: make(map[string]node.Interface)}

	// Set up test data
	hash := make(m.Hash)
	hash["node1"] = struct{}{}
	hash["node2"] = struct{}{}
	repo.data[prefixC2T+"test_hash"] = hash

	f := New(repo, factory)
	result := f.T("test_hash")

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(result))
	}
	if _, ok := result["node1"]; !ok {
		t.Error("expected node1 to be present")
	}
	if _, ok := result["node2"]; !ok {
		t.Error("expected node2 to be present")
	}
}

func TestFetcher_T_NoData(t *testing.T) {
	repo := &mockRepository{data: make(map[string]m.Hash)}
	factory := &mockFactory{created: make(map[string]node.Interface)}

	f := New(repo, factory)
	result := f.T("nonexistent_hash")

	if result != nil {
		t.Error("expected nil result for nonexistent hash")
	}
}

func TestFetcher_R(t *testing.T) {
	repo := &mockRepository{data: make(map[string]m.Hash)}
	factory := &mockFactory{created: make(map[string]node.Interface)}

	// Set up test data
	hash := make(m.Hash)
	hash["result1"] = struct{}{}
	hash["result2"] = struct{}{}
	repo.data[prefixC2R+"test_hash"] = hash

	f := New(repo, factory)
	result := f.R("test_hash")

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(result))
	}
	if _, ok := result["result1"]; !ok {
		t.Error("expected result1 to be present")
	}
	if _, ok := result["result2"]; !ok {
		t.Error("expected result2 to be present")
	}
}

func TestFetcher_R_NoData(t *testing.T) {
	repo := &mockRepository{data: make(map[string]m.Hash)}
	factory := &mockFactory{created: make(map[string]node.Interface)}

	f := New(repo, factory)
	result := f.R("nonexistent_hash")

	if result != nil {
		t.Error("expected nil result for nonexistent hash")
	}
}

func TestDefault(t *testing.T) {
	f := Default()
	if f == nil {
		t.Error("expected Default to return non-nil fetcher")
	}
}
