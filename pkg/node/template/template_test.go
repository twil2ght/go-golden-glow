package template

import (
	"goldenglow/m"
	"goldenglow/pkg/container/positioner"
	"goldenglow/pkg/node"
	"goldenglow/pkg/variable"
	"testing"
)

type mockRepo struct {
	items m.Hash
}

func (r *mockRepo) HGet(_ string) (m.Hash, error) {
	return r.items, nil
}

func (r *mockRepo) HSet(_ string, _ m.Hash) error {
	return nil
}

func (r *mockRepo) Get(_ string) (string, error) {
	return "", nil
}

func (r *mockRepo) Set(_, _ string) error {
	return nil
}

func (r *mockRepo) HDel(tag string, subKeys ...string) {}

func (r *mockRepo) Shutdown() error {
	return nil
}

func (r *mockRepo) Init() error {
	return nil
}

// createTestNode creates a test node using the real node factory
func createTestNode(value string) node.Interface {
	factory := node.DefaultFactory
	return factory.Create(value)
}
func createTestVarSet(kv m.Map[string]) variable.Set {
	if kv == nil {
		return nil
	}
	varSet := make(variable.Set)
	for k, v := range kv {
		varSet[k] = variable.New(k, v)
	}
	return varSet
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestCore_Get(t *testing.T) {
	testErr := &testError{msg: "test error"}

	tests := []struct {
		name        string
		templates   m.Hash
		targetValue string
		sourceErr   error
		wantErr     bool
		wantCount   int
		givenVars   map[string]string
	}{
		{
			name: "single matching template",
			templates: m.Hash{
				"hello $1": struct{}{},
			},
			targetValue: "hello world",
			wantErr:     false,
			wantCount:   1,
		},
		{
			name: "multiple matching templates",
			templates: m.Hash{
				"hello $1":       struct{}{},
				"hello $1 $2":    struct{}{},
				"$1 hello world": struct{}{},
			},
			targetValue: "hello world",
			wantErr:     false,
			wantCount:   1, // AllTemplates returns all matching templates (from general to specific)
		},
		{
			name:        "no matching templates",
			templates:   m.Hash{},
			targetValue: "hello world",
			wantErr:     false,
			wantCount:   0,
		},
		{
			name: "no matching pattern",
			templates: m.Hash{
				"goodbye $1": struct{}{},
			},
			targetValue: "hello world",
			wantErr:     false,
			wantCount:   0,
		},
		{
			name:        "source error",
			templates:   nil,
			targetValue: "hello world",
			sourceErr:   testErr,
			wantErr:     false,
			wantCount:   0,
		},
		{
			name: "single matching template v2",
			templates: m.Hash{
				"Zero says [input_Cs] $1 to Susie": struct{}{},
			},
			targetValue: "Zero says [input_Cs] zero says if Zero says $1 to Susie to Susie to Susie",
			wantErr:     false,
			wantCount:   1,
		},
		{
			name: "single matching template v3",
			templates: m.Hash{
				"Zero says [input_Cs] $1 to Susie": struct{}{},
			},
			targetValue: "Zero says [input_Cs] $1 to Susie",
			wantErr:     false,
			wantCount:   1,
		},
		{
			name: "real data test",
			templates: m.Hash{
				"Zero says if Zero says $1 to Susie to Susie": struct{}{},
			},
			targetValue: "Zero says if Zero says good morning to Susie to Susie",
			wantErr:     false,
			wantCount:   1,
		},
		{
			name: "real data test v2",
			templates: m.Hash{
				"$1 is greater than 18": struct{}{},
			},
			targetValue: "$1 is greater than $2",
			wantErr:     false,
			wantCount:   1,
			givenVars:   map[string]string{"$1": "20", "$2": "18"},
		},
		{
			name: "real data test v2",
			templates: m.Hash{
				"check if $1 is greater than $2": struct{}{},
			},
			targetValue: "check if $1 is greater than 5",
			wantErr:     false,
			wantCount:   1,
		},
		{
			name: "real data test v2",
			templates: m.Hash{
				"$1 is $2":           struct{}{},
				"$1 is less than 19": struct{}{},
				"$1 is less than 23": struct{}{},
			},
			targetValue: "$1 is less than $2",
			wantErr:     false,
			wantCount:   1,
			givenVars:   map[string]string{"$1": "18", "$2": "19"},
		},
		{
			name: "real data test v3",
			templates: m.Hash{
				"$1 is $2": struct{}{},
			},
			targetValue: "$1 is less than $2",
			wantErr:     false,
			wantCount:   1,
			givenVars:   map[string]string{"$1": "18", "$2": "19"},
		},
		{
			name: "real data test v3",
			templates: m.Hash{
				"[node:executor] [namespace:builder] [mode:multi_condition] [type:input] [value:$1]": struct{}{},
			},
			targetValue: "[node:executor] [namespace:builder] [mode:multi_condition] [type:input] [value:$1]",
			wantErr:     false,
			wantCount:   1,
			givenVars:   map[string]string{"$1": "18"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := New(&mockRepo{items: tt.templates}, node.DefaultFactory, positioner.Default())
			core.BanFilter()
			if core == nil {
				t.Fatalf("New() failed")
			}

			target := createTestNode(tt.targetValue)
			varSet := createTestVarSet(tt.givenVars)
			state := node.GenVariableState(varSet)
			target.VarSetRegistry().Register(state, varSet)
			result := core.GetTemplate(target, state)
			if (result == nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", "", tt.wantErr)
				return
			}
			if !tt.wantErr && len(result) != tt.wantCount {
				t.Errorf("Get() returned %d matches, want %d", len(result), tt.wantCount)
			}
			// Print all matched templates
			if !tt.wantErr && len(result) > 0 {
				t.Logf("Matched templates for target %q:", tt.targetValue)
				for key, n := range result {
					t.Logf("  - %q -> value: %q", key, n.Value())
				}
			}
		})
	}
}
