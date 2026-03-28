package executor

import (
	"testing"

	"goldenglow/node"
)

// mockNode creates a test node with the given value
func mockNode(value string) *BaseNode {
	factory := node.DefaultFactory()
	n, _ := factory.New(value)
	return &BaseNode{
		Base:     *n.(*node.Base),
		handlers: make(map[string]Handler),
	}
}

func TestBaseNode_GetParams_Simple(t *testing.T) {
	n := mockNode("[name:John]")
	params := n.GetParams()

	if len(params) != 1 {
		t.Errorf("expected 1 param, got %d", len(params))
	}
	if params["name"] != "John" {
		t.Errorf("expected name=John, got name=%s", params["name"])
	}
}

func TestBaseNode_GetParams_Multiple(t *testing.T) {
	n := mockNode("[name:John] [age:30] [city:NYC]")
	params := n.GetParams()

	if len(params) != 3 {
		t.Errorf("expected 3 params, got %d", len(params))
	}
	if params["name"] != "John" {
		t.Errorf("expected name=John, got name=%s", params["name"])
	}
	if params["age"] != "30" {
		t.Errorf("expected age=30, got age=%s", params["age"])
	}
	if params["city"] != "NYC" {
		t.Errorf("expected city=NYC, got city=%s", params["city"])
	}
}

func TestBaseNode_GetParams_NestedBrackets(t *testing.T) {
	n := mockNode("[expr:[a]+[b]]")
	params := n.GetParams()

	if len(params) != 1 {
		t.Errorf("expected 1 param, got %d", len(params))
	}
	if params["expr"] != "[a]+[b]" {
		t.Errorf("expected expr=[a]+[b], got expr=%s", params["expr"])
	}
}

func TestBaseNode_GetParams_DeepNesting(t *testing.T) {
	n := mockNode("[config:[key:[nested:value]]]")
	params := n.GetParams()

	if len(params) != 1 {
		t.Errorf("expected 1 param, got %d", len(params))
	}
	if params["config"] != "[key:[nested:value]]" {
		t.Errorf("expected config=[key:[nested:value]], got config=%s", params["config"])
	}
}

func TestBaseNode_GetParams_ArrayValue(t *testing.T) {
	n := mockNode("[arr:[1,2,3]]")
	params := n.GetParams()

	if len(params) != 1 {
		t.Errorf("expected 1 param, got %d", len(params))
	}
	if params["arr"] != "[1,2,3]" {
		t.Errorf("expected arr=[1,2,3], got arr=%s", params["arr"])
	}
}

func TestBaseNode_GetParams_WithSpaces(t *testing.T) {
	n := mockNode("[ name : John Doe ]")
	params := n.GetParams()

	if params["name"] != "John Doe" {
		t.Errorf("expected name='John Doe', got name=%s", params["name"])
	}
}

func TestBaseNode_GetParams_EmptyValue(t *testing.T) {
	n := mockNode("[key:]")
	params := n.GetParams()

	if params["key"] != "" {
		t.Errorf("expected empty value, got %s", params["key"])
	}
}

func TestBaseNode_GetParams_EmptyKey(t *testing.T) {
	n := mockNode("[:value]")
	params := n.GetParams()

	if _, exists := params[""]; exists {
		t.Error("expected empty key to be ignored")
	}
}

func TestBaseNode_GetParams_NoColon(t *testing.T) {
	n := mockNode("[no colon here]")
	params := n.GetParams()

	if len(params) != 0 {
		t.Errorf("expected 0 params for invalid format, got %d", len(params))
	}
}

func TestBaseNode_GetParams_NoBrackets(t *testing.T) {
	n := mockNode("just plain text")
	params := n.GetParams()

	if len(params) != 0 {
		t.Errorf("expected 0 params for text without brackets, got %d", len(params))
	}
}

func TestBaseNode_GetParams_MixedContent(t *testing.T) {
	n := mockNode("some text [key:value] more text [another:123]")
	params := n.GetParams()

	if len(params) != 2 {
		t.Errorf("expected 2 params, got %d", len(params))
	}
	if params["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", params["key"])
	}
	if params["another"] != "123" {
		t.Errorf("expected another=123, got another=%s", params["another"])
	}
}

func TestBaseNode_GetParams_UnclosedBracket(t *testing.T) {
	n := mockNode("[unclosed:value")
	params := n.GetParams()

	if len(params) != 0 {
		t.Errorf("expected 0 params for unclosed bracket, got %d", len(params))
	}
}

func TestBaseNode_GetParams_UnopenedBracket(t *testing.T) {
	n := mockNode("unopened:value]")
	params := n.GetParams()

	// The value after ] might be parsed depending on position
	// This test mainly ensures no panic occurs
	_ = params
}

func TestBaseNode_GetParams_NamespaceKey(t *testing.T) {
	n := mockNode("[namespace:test_plugin]")
	params := n.GetParams()

	if params[KeyNamespace] != "test_plugin" {
		t.Errorf("expected namespace=test_plugin, got %s", params[KeyNamespace])
	}
}

func TestBaseNode_GetParams_ComplexNested(t *testing.T) {
	n := mockNode("[data:[type:json][content:[key1:val1][key2:val2]]]")
	params := n.GetParams()

	if len(params) != 1 {
		t.Errorf("expected 1 param, got %d", len(params))
	}
	expected := "[type:json][content:[key1:val1][key2:val2]]"
	if params["data"] != expected {
		t.Errorf("expected data=%s, got data=%s", expected, params["data"])
	}
}

func TestBaseNode_GetParams_BracketInMiddle(t *testing.T) {
	n := mockNode("prefix[key:value]suffix")
	params := n.GetParams()

	if len(params) != 1 {
		t.Errorf("expected 1 param, got %d", len(params))
	}
	if params["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", params["key"])
	}
}

func TestBaseNode_GetParams_MultipleColons(t *testing.T) {
	n := mockNode("[time:12:30:45]")
	params := n.GetParams()

	if len(params) != 1 {
		t.Errorf("expected 1 param, got %d", len(params))
	}
	if params["time"] != "12:30:45" {
		t.Errorf("expected time=12:30:45, got time=%s", params["time"])
	}
}

func TestBaseNode_GetParams_NestedWithMultipleColons(t *testing.T) {
	n := mockNode("[range:[start:10:00][end:12:00]]")
	params := n.GetParams()

	if len(params) != 1 {
		t.Errorf("expected 1 param, got %d", len(params))
	}
	expected := "[start:10:00][end:12:00]"
	if params["range"] != expected {
		t.Errorf("expected range=%s, got range=%s", expected, params["range"])
	}
}

func TestBaseNode_Execute_NoNamespace(t *testing.T) {
	n := mockNode("[key:value]")
	n.handlers = make(map[string]Handler)
	err := n.Execute()

	if err == nil {
		t.Error("expected error when namespace is missing")
	}
}

func TestBaseNode_Execute_UnregisteredPlugin(t *testing.T) {
	n := mockNode("[namespace:unregistered]")
	n.handlers = make(map[string]Handler)
	err := n.Execute()

	if err == nil {
		t.Error("expected error for unregistered plugin")
	}
}

func TestBaseNode_Execute_Success(t *testing.T) {
	executed := false
	n := mockNode("[namespace:test]")
	n.handlers = map[string]Handler{
		"test": func(params Parameters) error {
			executed = true
			return nil
		},
	}
	err := n.Execute()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !executed {
		t.Error("handler was not executed")
	}
}
