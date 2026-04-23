package variable

import (
	"testing"
)

func TestToRawText_SimpleReplacement(t *testing.T) {
	variables := Set{
		"$1": New("$1", "hello"),
		"$2": New("$2", "world"),
	}

	result, err := ToRawText("$1 $2", variables, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", result)
	}
}

func TestToRawText_NestedVariable(t *testing.T) {
	variables := Set{
		"$1": New("$1", "hello $2"),
		"$2": New("$2", "world"),
	}

	result, err := ToRawText("$1", variables, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", result)
	}
}

func TestToRawText_DeepNesting(t *testing.T) {
	variables := Set{
		"$1": New("$1", "a $2"),
		"$2": New("$2", "b $3"),
		"$3": New("$3", "c"),
	}

	result, err := ToRawText("$1", variables, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "a b c" {
		t.Errorf("expected 'a b c', got '%s'", result)
	}
}

func TestToRawText_CycleDetection(t *testing.T) {
	variables := Set{
		"$1": New("$1", "$2"),
		"$2": New("$2", "$1"),
	}

	// Should not hang - cycle should be detected and stopped
	result, err := ToRawText("$1", variables, false)
	if err == nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// When cycle is detected, the variable should be left as-is
	if result != "$1" && result != "$2" {
		t.Errorf("expected cycle to be detected, got '%s'", result)
	}
}

func TestToRawText_SelfReference(t *testing.T) {
	variables := Set{
		"$1": New("$1", "value $1"),
	}

	// Self-reference should be detected as cycle
	_, err := ToRawText("$1", variables, false)
	if err == nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should stop at first replacement when cycle detected
}

func TestToRawText_TriangleCycle(t *testing.T) {
	variables := Set{
		"$1": New("$1", "$2"),
		"$2": New("$2", "$3"),
		"$3": New("$3", "$1"),
	}

	// Triangle cycle: $1 -> $2 -> $3 -> $1
	result, err := ToRawText("$1", variables, false)
	if err == nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should detect cycle and stop
	if result != "$1" && result != "$2" && result != "$3" {
		t.Errorf("expected cycle to be detected, got '%s'", result)
	}
}

func TestToRawText_StrictMode_MissingVariable(t *testing.T) {
	variables := Set{
		"$1": New("$1", "hello"),
	}

	_, err := ToRawText("$1 $2", variables, true)
	if err == nil {
		t.Error("expected error in strict mode for missing variable")
	}
}

func TestToRawText_NonStrictMode_MissingVariable(t *testing.T) {
	variables := Set{
		"$1": New("$1", "hello"),
	}

	result, err := ToRawText("$1 $2", variables, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello $2" {
		t.Errorf("expected 'hello $2', got '%s'", result)
	}
}

func TestToRawText_NoVariables(t *testing.T) {
	variables := Set{}

	result, err := ToRawText("hello world", variables, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", result)
	}
}

func TestToRawText_EmptyString(t *testing.T) {
	variables := Set{
		"$1": New("$1", "hello"),
	}

	result, err := ToRawText("", variables, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestToRawText_MultipleSameVariable(t *testing.T) {
	variables := Set{
		"$1": New("$1", "hi"),
	}

	result, err := ToRawText("$1 $1 $1", variables, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hi hi hi" {
		t.Errorf("expected 'hi hi hi', got '%s'", result)
	}
}

func TestToRawText_EmptyVariableValue(t *testing.T) {
	variables := Set{
		"$1": New("$1", ""),
	}

	result, err := ToRawText("hello $1 world", variables, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello  world" {
		t.Errorf("expected 'hello  world', got '%s'", result)
	}
}

func TestToRawText_ComplexMixedContent(t *testing.T) {
	variables := Set{
		"$1": New("$1", "quick"),
		"$2": New("$2", "brown"),
		"$3": New("$3", "fox"),
	}

	result, err := ToRawText("The $1 $2 $3 jumps", variables, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "The quick brown fox jumps" {
		t.Errorf("expected 'The quick brown fox jumps', got '%s'", result)
	}
}

func TestToRawText_PartialCycle(t *testing.T) {
	// $1 references $2, but $2 is a normal value
	// This should work fine, not a cycle
	variables := Set{
		"$1": New("$1", "hello $2"),
		"$2": New("$2", "world"),
	}

	result, err := ToRawText("$1", variables, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", result)
	}
}

func TestToRawText_CycleWithText(t *testing.T) {
	// Cycle with additional text
	variables := Set{
		"$1": New("$1", "prefix $2 suffix"),
		"$2": New("$2", "$1"),
	}

	result, err := ToRawText("$1", variables, false)
	if err == nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should detect cycle and preserve the structure
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestBase_Copy(t *testing.T) {
	original := New("key1", "value1")
	copied := original.Copy()

	if copied.Name() != original.Name() {
		t.Errorf("copied name mismatch: expected '%s', got '%s'", original.Name(), copied.Name())
	}
	if copied.Value() != original.Value() {
		t.Errorf("copied value mismatch: expected '%s', got '%s'", original.Value(), copied.Value())
	}

	// Modify copy should not affect original
	copied.Set("newvalue")
	if original.Value() != "value1" {
		t.Error("modifying copy affected original")
	}
}

func TestCopy_Set(t *testing.T) {
	original := Set{
		"$1": New("$1", "hello"),
		"$2": New("$2", "world"),
	}

	copied := Copy(original)

	if len(copied) != len(original) {
		t.Errorf("copied set size mismatch: expected %d, got %d", len(original), len(copied))
	}

	// Modify copy should not affect original
	copied["$1"].Set("modified")
	if original["$1"].Value() != "hello" {
		t.Error("modifying copy affected original set")
	}
}

func TestBase_OK(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"hello", true},
		{"", false},
		{"non-empty", true},
	}

	for _, tt := range tests {
		item := New("test", tt.value)
		if item.OK() != tt.want {
			t.Errorf("OK() for value '%s': expected %v, got %v", tt.value, tt.want, item.OK())
		}
	}
}

func TestBase_Set_EmptyValue(t *testing.T) {
	item := New("test", "initial")
	err := item.Set("")
	if err == nil {
		t.Error("expected error when setting empty value")
	}
	if item.Value() != "initial" {
		t.Error("value should not change when Set returns error")
	}
}

func TestNew_EmptyKey(t *testing.T) {
	item := New("", "value")
	if item.Name() != "you get an empty key" {
		t.Errorf("expected default key name, got '%s'", item.Name())
	}
}

func TestSet_HasCycle_NoCycle(t *testing.T) {
	variables := Set{
		"$1": New("$1", "hello $2"),
		"$2": New("$2", "world"),
	}
	if variables.HasCycle() {
		t.Error("expected no cycle, but HasCycle returned true")
	}
}

func TestSet_HasCycle_SelfReference(t *testing.T) {
	variables := Set{
		"$1": New("$1", "value $1"),
	}
	if !variables.HasCycle() {
		t.Error("expected cycle for self-reference, but HasCycle returned false")
	}
}

func TestSet_HasCycle_DirectCycle(t *testing.T) {
	variables := Set{
		"$1": New("$1", "$2"),
		"$2": New("$2", "$1"),
	}
	if !variables.HasCycle() {
		t.Error("expected cycle for $1 <-> $2, but HasCycle returned false")
	}
}

func TestSet_HasCycle_TriangleCycle(t *testing.T) {
	variables := Set{
		"$1": New("$1", "$2"),
		"$2": New("$2", "$3"),
		"$3": New("$3", "$1"),
	}
	if !variables.HasCycle() {
		t.Error("expected cycle for triangle $1 -> $2 -> $3 -> $1, but HasCycle returned false")
	}
}

func TestSet_HasCycle_LongChainNoCycle(t *testing.T) {
	variables := Set{
		"$1": New("$1", "a $2"),
		"$2": New("$2", "b $3"),
		"$3": New("$3", "c $4"),
		"$4": New("$4", "d"),
	}
	if variables.HasCycle() {
		t.Error("expected no cycle for linear chain, but HasCycle returned true")
	}
}

func TestSet_HasCycle_MultipleIndependentCycles(t *testing.T) {
	variables := Set{
		"$1": New("$1", "$2"),
		"$2": New("$2", "$1"),
		"$3": New("$3", "$4"),
		"$4": New("$4", "$3"),
	}
	if !variables.HasCycle() {
		t.Error("expected cycle detection with multiple independent cycles")
	}
}

func TestSet_HasCycle_ComplexNestedNoCycle(t *testing.T) {
	variables := Set{
		"$1": New("$1", "$2 $3"),
		"$2": New("$2", "hello"),
		"$3": New("$3", "$4 world"),
		"$4": New("$4", "beautiful"),
	}
	if variables.HasCycle() {
		t.Error("expected no cycle for complex nested structure")
	}
}

func TestSet_HasCycle_EmptySet(t *testing.T) {
	variables := Set{}
	if variables.HasCycle() {
		t.Error("expected no cycle for empty set")
	}
}

func TestSet_HasCycle_NoVariableReferences(t *testing.T) {
	variables := Set{
		"$1": New("$1", "hello world"),
		"$2": New("$2", "no refs here"),
	}
	if variables.HasCycle() {
		t.Error("expected no cycle when no variable references exist")
	}
}

func TestSet_HasCycle_ReferenceToNonExistent(t *testing.T) {
	variables := Set{
		"$1": New("$1", "$999"),
	}
	if variables.HasCycle() {
		t.Error("expected no cycle when referencing non-existent variable")
	}
}
