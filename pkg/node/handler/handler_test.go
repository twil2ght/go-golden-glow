package handler

import (
	"goldenglow/m"
	"goldenglow/pkg/registry"
	"goldenglow/pkg/variable"
	"testing"
)

func TestNew(t *testing.T) {
	base := New[string]("test_value", registry.New[string]())
	if base.Interface == nil {
		t.Error("expected Interface to be initialized")
	}
	if base.handlers == nil {
		t.Error("expected handlers to be initialized")
	}
	if base.Value() != "test_value" {
		t.Errorf("expected value 'test_value', got '%s'", base.Value())
	}
}

func TestBase_Handlers(t *testing.T) {
	base := New[string]("test", registry.New[string]())
	handlers := base.Handlers()
	if handlers == nil {
		t.Error("expected handlers to be non-nil")
	}
	if handlers.Len() != 0 {
		t.Errorf("expected empty handlers, got len=%d", handlers.Len())
	}
}

func TestExecutor_Execute(t *testing.T) {
	exec := &executor{
		Base: New[ExecuteHandler]("test [namespace:test_handler]", registry.New[ExecuteHandler]()),
	}

	called := false
	exec.handlers.Register("test_handler", func(params Parameters) {
		called = true
		val, _ := params.Get("namespace")
		if val != "test_handler" {
			t.Errorf("expected namespace 'test_handler', got '%s'", val)
		}
	})

	exec.Execute("some_state")

	if !called {
		t.Error("expected handler to be called")
	}
}

func TestExecutor_Execute_NoHandler(t *testing.T) {
	exec := &executor{
		Base: New[ExecuteHandler]("test [namespace:unknown]", registry.New[ExecuteHandler]()),
	}

	// No handler registered, should not panic
	exec.Execute("some_state")
}

func TestChecker_Check(t *testing.T) {
	check := &checker{
		Base: New[CheckHandler]("test [namespace:test_handler]", registry.New[CheckHandler]()),
	}

	check.handlers.Register("test_handler", func(params Parameters) bool {
		val, _ := params.Get("namespace")
		return val == "test_handler"
	})

	result := check.Check("some_state")
	if !result {
		t.Error("expected check to return true")
	}
}

func TestChecker_Check_False(t *testing.T) {
	check := &checker{
		Base: New[CheckHandler]("test [namespace:test_handler]", registry.New[CheckHandler]()),
	}

	check.handlers.Register("test_handler", func(params Parameters) bool {
		return false
	})

	result := check.Check("some_state")
	if result {
		t.Error("expected check to return false")
	}
}

func TestChecker_Check_NoHandler(t *testing.T) {
	check := &checker{
		Base: New[CheckHandler]("test [namespace:unknown]", registry.New[CheckHandler]()),
	}

	result := check.Check("some_state")
	if result {
		t.Error("expected check to return false when no handler")
	}
}

func TestExtractor_KeyDist(t *testing.T) {
	ext := &extractor{
		Base: New[ExtractorHandler]("test [dist:test_dist]", registry.New[ExtractorHandler]()),
	}

	dist := ext.KeyDist()
	if dist != "test_dist" {
		t.Errorf("expected dist 'test_dist', got '%s'", dist)
	}
}

func TestExtractor_KeyDist_NoDist(t *testing.T) {
	ext := &extractor{
		Base: New[ExtractorHandler]("test", registry.New[ExtractorHandler]()),
	}

	dist := ext.KeyDist()
	if dist != "" {
		t.Errorf("expected empty dist, got '%s'", dist)
	}
}

func TestExtractor_Extract(t *testing.T) {
	ext := &extractor{
		Base: New[ExtractorHandler]("test [namespace:test_handler]", registry.New[ExtractorHandler]()),
	}

	ext.handlers.Register("test_handler", func(params Parameters) variable.ValueMap {
		return variable.NewValueMap(m.Hash{"test_var": struct{}{}, "test_value": struct{}{}})
	})

	result := ext.Extract("some_state")
	if result == nil {
		t.Error("expected non-nil ValueMap")
	}
	hash := result.Get()
	if hash == nil {
		t.Error("expected non-nil hash")
	}
}

func TestExtractor_Extract_NoHandler(t *testing.T) {
	ext := &extractor{
		Base: New[ExtractorHandler]("test [namespace:unknown]", registry.New[ExtractorHandler]()),
	}

	result := ext.Extract("some_state")
	if result != nil {
		t.Error("expected nil ValueMap when no handler")
	}
}

func TestGetParameters(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]string
	}{
		{"", map[string]string{}},
		{"[key:value]", map[string]string{"key": "value"}},
		{"[key1:value1] [key2:value2]", map[string]string{"key1": "value1", "key2": "value2"}},
		{"text [key:value] more", map[string]string{"key": "value"}},
		{"[namespace:test] [dist:output]", map[string]string{"namespace": "test", "dist": "output"}},
	}

	for _, tt := range tests {
		params := GetParameters(tt.input)
		for k, v := range tt.expected {
			val, err := params.Get(k)
			if err != nil {
				t.Errorf("expected key '%s' to exist", k)
			}
			if val != v {
				t.Errorf("for key '%s', expected '%s', got '%s'", k, v, val)
			}
		}
		if params.Len() != len(tt.expected) {
			t.Errorf("expected %d params, got %d", len(tt.expected), params.Len())
		}
	}
}
