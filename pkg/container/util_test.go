package container

import (
	"goldenglow/pkg/variable"
	"reflect"
	"testing"

	"goldenglow/m"
	"goldenglow/pkg/node"
	"goldenglow/pkg/registry"
)

func TestGetCompatibleSets(t *testing.T) {
	// Create a node
	n := node.New("test")

	// Cast to *node.Node to access private field for testing
	testNode := n.(*node.Node)

	// Test 1: Empty registry
	sets, ok := getCompatibleSets(n, variable.Set{})
	if !ok || sets != nil {
		t.Errorf("Expected (nil, true) for empty registry, got (%v, %v)", sets, ok)
	}

	// Create some variable sets
	set1 := variable.Set{
		"a": variable.New("a", "1"),
		"b": variable.New("b", "2"),
	}
	set2 := variable.Set{
		"a": variable.New("a", "1"),
		"c": variable.New("c", "3"),
	}
	set3 := variable.Set{
		"a": variable.New("a", "4"), // incompatible
	}

	// Register them
	reg := testNode.VarSetRegistry().(*registry.DefaultRegistry[variable.Set])
	reg.Register("set1", set1)
	reg.Register("set2", set2)
	reg.Register("set3", set3)

	// Test 2: Base set that matches set1 and set2
	base := variable.Set{
		"a": variable.New("a", "1"),
	}

	sets, ok = getCompatibleSets(n, base)
	if !ok || len(sets) != 2 {
		t.Errorf("Expected 2 compatible sets, got %d, ok=%v", len(sets), ok)
	}
	expected := []variable.Set{set1, set2}
	if !reflect.DeepEqual(sets, expected) {
		t.Errorf("Expected %v, got %v", expected, sets)
	}

	// Test 3: Base set that matches none
	base2 := variable.Set{
		"a": variable.New("a", "5"),
	}

	sets, ok = getCompatibleSets(n, base2)
	if ok || sets != nil {
		t.Errorf("Expected (nil, false) for no matches, got (%v, %v)", sets, ok)
	}

	// Test 4: Empty base set - all should be compatible
	sets, ok = getCompatibleSets(n, variable.Set{})
	if !ok || len(sets) != 3 {
		t.Errorf("Expected 3 compatible sets for empty base, got %d, ok=%v", len(sets), ok)
	}
}

func TestMergeVariables(t *testing.T) {
	// Create exclude node
	excludeNode := node.New("exclude")

	// Create other nodes
	node1 := node.New("node1")
	node2 := node.New("node2")
	node3 := node.New("node3")

	// Cast to access registries
	n1 := node1.(*node.Node)
	n2 := node2.(*node.Node)
	n3 := node3.(*node.Node)

	// Create variable sets
	set1 := variable.Set{
		"a": variable.New("a", "1"),
	}
	set2 := variable.Set{
		"a": variable.New("a", "1"),
		"b": variable.New("b", "2"),
	}
	set3 := variable.Set{
		"a": variable.New("a", "1"),
		"c": variable.New("c", "3"),
	}

	// Register sets
	reg1 := n1.VarSetRegistry().(*registry.DefaultRegistry[variable.Set])
	reg1.Register("s1", set1)

	reg2 := n2.VarSetRegistry().(*registry.DefaultRegistry[variable.Set])
	reg2.Register("s2", set2)

	reg3 := n3.VarSetRegistry().(*registry.DefaultRegistry[variable.Set])
	reg3.Register("s3", set3)

	// Create ts map
	ts := m.Map[node.Interface]{
		"exclude": excludeNode,
		"node1":   node1,
		"node2":   node2,
		"node3":   node3,
	}

	// Initial varSet
	varSet := variable.Set{
		"a": variable.New("a", "1"),
	}

	// Test 1: Successful merge
	ok := mergeVariables(excludeNode, ts, varSet)
	if !ok {
		t.Errorf("Expected merge to succeed, but it failed")
	}
	// Check if b and c were added
	if _, exists := varSet["b"]; !exists {
		t.Errorf("Expected 'b' to be added to varSet")
	}
	if _, exists := varSet["c"]; !exists {
		t.Errorf("Expected 'c' to be added to varSet")
	}

	// Test 2: Incompatible sets are skipped, merge still succeeds
	node4 := node.New("node4")
	n4 := node4.(*node.Node)
	reg4 := n4.VarSetRegistry().(*registry.DefaultRegistry[variable.Set])
	reg4.Register("incompatible", variable.Set{
		"a": variable.New("a", "2"), // incompatible with base
	})
	ts["node4"] = node4

	varSet2 := variable.Set{
		"a": variable.New("a", "1"),
	}
	ok = mergeVariables(excludeNode, ts, varSet2)
	if !ok {
		t.Errorf("Expected merge to succeed by skipping incompatible node, but it failed")
	}

	// Test 3: No nodes besides exclude
	tsEmpty := m.Map[node.Interface]{
		"exclude": excludeNode,
	}
	varSet3 := variable.Set{}
	ok = mergeVariables(excludeNode, tsEmpty, varSet3)
	if !ok {
		t.Errorf("Expected true for no other nodes, got false")
	}
}

func TestMergeVariablesMultiSets(t *testing.T) {
	// Create exclude node
	excludeNode := node.New("exclude")

	// Create nodes
	node1 := node.New("node1")
	node2 := node.New("node2")

	// Cast
	n1 := node1.(*node.Node)
	n2 := node2.(*node.Node)

	// Node1 has one set
	set1 := variable.Set{
		"a": variable.New("a", "1"),
	}
	reg1 := n1.VarSetRegistry().(*registry.DefaultRegistry[variable.Set])
	reg1.Register("s1", set1)

	// Node2 has multiple sets
	set2a := variable.Set{
		"a": variable.New("a", "1"),
		"b": variable.New("b", "2"),
	}
	set2b := variable.Set{
		"a": variable.New("a", "1"),
		"c": variable.New("c", "3"),
	}
	reg2 := n2.VarSetRegistry().(*registry.DefaultRegistry[variable.Set])
	reg2.Register("s2a", set2a)
	reg2.Register("s2b", set2b)

	// ts map
	ts := m.Map[node.Interface]{
		"exclude": excludeNode,
		"node1":   node1,
		"node2":   node2,
	}

	// varSet
	varSet := variable.Set{
		"a": variable.New("a", "1"),
	}

	// Merge
	ok := mergeVariables(excludeNode, ts, varSet)
	if !ok {
		t.Errorf("Expected merge to succeed with multi-sets")
	}

	// Check that variables from one of the sets are added
	// Since it picks the first merged set, it should have either b or c
	hasB := false
	hasC := false
	if _, exists := varSet["b"]; exists {
		hasB = true
	}
	if _, exists := varSet["c"]; exists {
		hasC = true
	}
	if !hasB && !hasC {
		t.Errorf("Expected at least one of 'b' or 'c' to be added")
	}
}
