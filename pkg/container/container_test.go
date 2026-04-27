package container

import (
	"goldenglow/m"
	"goldenglow/pkg/node"
	"goldenglow/pkg/registry"
	"goldenglow/pkg/variable"
	"testing"
)

// mockFetcher implements fetcher.Interface for testing
type mockFetcher struct {
	tData m.Map[node.Interface]
	rData m.Map[node.Interface]
}

func (mf *mockFetcher) T(hash string) m.Map[node.Interface] {
	return mf.tData
}

func (mf *mockFetcher) R(hash string) m.Map[node.Interface] {
	return mf.rData
}

// mockNode implements node.Interface for testing
type mockNode struct {
	value string
}

func (mn *mockNode) Activate() {

}

func (mn *mockNode) IsActivated() bool {
	return true
}

func (mn *mockNode) Execute(state string)                 {}
func (mn *mockNode) Value() string                        { return mn.value }
func (mn *mockNode) VarKeys() []string                    { return nil }
func (mn *mockNode) ToTextWithNoVars(state string) string { return mn.value }
func (mn *mockNode) VarSetRegistry() registry.Interface[variable.Set] {
	return registry.New[variable.Set]()
}

func TestNew(t *testing.T) {
	f := &mockFetcher{}
	c := New("test_hash", f)
	if c == nil {
		t.Error("expected New to return non-nil container")
	}
}

func TestNewWithDefaultFetcher(t *testing.T) {
	c := NewWithDefaultFetcher("test_hash")
	if c == nil {
		t.Error("expected NewWithDefaultFetcher to return non-nil container")
	}
}

func TestContainer_T(t *testing.T) {
	f := &mockFetcher{
		tData: m.Map[node.Interface]{
			"trigger1": &mockNode{value: "trigger1"},
		},
	}
	c := New("test_hash", f)
	c.(*container).t = f.tData // Simulate fetch

	result := c.T()
	if len(result) != 1 {
		t.Errorf("expected 1 trigger node, got %d", len(result))
	}
	if _, ok := result["trigger1"]; !ok {
		t.Error("expected trigger1 to be present")
	}
}

func TestContainer_R(t *testing.T) {
	f := &mockFetcher{
		rData: m.Map[node.Interface]{
			"result1": &mockNode{value: "result1"},
		},
	}
	c := New("test_hash", f)
	c.(*container).r = f.rData // Simulate fetch
	c.(*container).s = m.Map[[]string]{
		"result1": {"state1", "state2"},
	}

	rMap, sMap := c.R()
	if len(rMap) != 1 {
		t.Errorf("expected 1 result node, got %d", len(rMap))
	}
	if _, ok := rMap["result1"]; !ok {
		t.Error("expected result1 to be present")
	}
	if len(sMap) != 1 {
		t.Errorf("expected 1 state entry, got %d", len(sMap))
	}
	if states, ok := sMap["result1"]; !ok || len(states) != 2 {
		t.Errorf("expected 2 states for result1, got %v", states)
	}
}

func TestContainer_Forward_FetchFails(t *testing.T) {
	f := &mockFetcher{} // No data, fetch will fail
	c := New("test_hash", f)
	tNode := &mockNode{value: "test"}

	result := c.Forward(tNode, "state")
	if result {
		t.Error("expected Forward to return false when fetch fails")
	}
}

func TestContainer_Forward_Success(t *testing.T) {
	f := &mockFetcher{
		tData: m.Map[node.Interface]{
			"trigger1": &mockNode{value: "trigger1"},
		},
		rData: m.Map[node.Interface]{
			"result1": &mockNode{value: "result1"},
		},
	}
	c := New("test_hash", f)
	tNode := &mockNode{value: "test"}

	// Set up varSet
	c.(*container).varSet = variable.Set{
		"$1": variable.New("$1", "value1"),
	}

	result := c.Forward(tNode, "state")
	if !result {
		t.Error("expected Forward to return true for successful processing")
	}
}
