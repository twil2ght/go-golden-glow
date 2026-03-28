package template

import (
	"goldenglow/storage"
	"regexp"
	"testing"

	"goldenglow/node"
	"goldenglow/variable"
)

// mockSource is a test implementation of the Source interface
type mockSource struct {
	templates node.Set
	err       error
}

func (m *mockSource) GetTemplates() (node.Set, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.templates, nil
}

// createTestNode creates a test node using the real node factory
func createTestNode(value string) node.Item {
	factory := node.DefaultFactory()
	n, _ := factory.New(value)
	return n
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		source  Source
		varReg  *regexp.Regexp
		wantErr bool
	}{
		{
			name:    "valid initialization",
			source:  &mockSource{},
			varReg:  variable.VarReg,
			wantErr: false,
		},
		{
			name:    "nil source",
			source:  nil,
			varReg:  variable.VarReg,
			wantErr: true,
		},
		{
			name:    "nil varReg",
			source:  &mockSource{},
			varReg:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, err := New(tt.source, tt.varReg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && core == nil {
				t.Error("New() returned nil core without error")
			}
		})
	}
}

func TestCore_Get(t *testing.T) {
	testErr := &testError{msg: "test error"}

	tests := []struct {
		name      string
		templates node.Set
		target    node.Item
		sourceErr error
		wantErr   bool
		wantCount int
	}{
		{
			name: "single matching template",
			templates: node.Set{
				"hello $1": createTestNode("hello $1"),
			},
			target:    createTestNode("hello world"),
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "multiple matching templates",
			templates: node.Set{
				"hello $1":       createTestNode("hello $1"),
				"hello $1 $2":    createTestNode("hello $1 $2"),
				"$1 hello world": createTestNode("$1 hello world"),
			},
			target:    createTestNode("hello world"),
			wantErr:   false,
			wantCount: 1, // Most specific match should be kept
		},
		{
			name:      "no matching templates",
			templates: node.Set{},
			target:    createTestNode("hello world"),
			wantErr:   true,
			wantCount: 0,
		},
		{
			name: "no matching pattern",
			templates: node.Set{
				"goodbye $1": createTestNode("goodbye $1"),
			},
			target:    createTestNode("hello world"),
			wantErr:   true,
			wantCount: 0,
		},
		{
			name:      "source error",
			templates: nil,
			target:    createTestNode("hello world"),
			sourceErr: testErr,
			wantErr:   true,
			wantCount: 0,
		},
		{
			name: "single matching template v2",
			templates: node.Set{
				"Zero says [input_Cs] $1 to Susie": createTestNode("Zero says [input_Cs] $1 to Susie"),
			},
			target:    createTestNode("Zero says [input_Cs] zero says if Zero says $1 to Susie to Susie to Susie"),
			wantErr:   false,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := &mockSource{
				templates: tt.templates,
				err:       tt.sourceErr,
			}
			core, err := New(source, variable.VarReg)
			if err != nil {
				t.Fatalf("New() failed: %v", err)
			}

			result, err := core.Get(tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(result) != tt.wantCount {
				t.Errorf("Get() returned %d matches, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestCore_Match(t *testing.T) {
	source := &mockSource{templates: node.Set{}}
	core, err := New(source, variable.VarReg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	tests := []struct {
		name   string
		input  string
		target string
		want   bool
	}{
		{
			name:   "exact match",
			input:  "hello world",
			target: "hello world",
			want:   true,
		},
		{
			name:   "variable match",
			input:  "hello world",
			target: "hello $1",
			want:   true,
		},
		{
			name:   "multiple variables",
			input:  "hello beautiful world",
			target: "hello $1 $2",
			want:   true,
		},
		{
			name:   "no match",
			input:  "hello world",
			target: "goodbye $1",
			want:   false,
		},
		{
			name:   "empty target",
			input:  "hello world",
			target: "",
			want:   false,
		},
		{
			name:   "input shorter than target",
			input:  "hi",
			target: "hello world",
			want:   false,
		},
		{
			name:   "variable at start",
			input:  "john says hello",
			target: "$1 says hello",
			want:   true,
		},
		{
			name:   "variable at end",
			input:  "say hello to mary",
			target: "say hello to $1",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := core.Match(tt.input, tt.target)
			if got != tt.want {
				t.Errorf("Match(%q, %q) = %v, want %v", tt.input, tt.target, got, tt.want)
			}
		})
	}
}

func TestCore_segment(t *testing.T) {
	source := &mockSource{templates: node.Set{}}
	c := &core{
		source: source,
		varReg: variable.VarReg,
	}

	tests := []struct {
		name string
		tpl  string
		want []string
	}{
		{
			name: "no variables",
			tpl:  "hello world",
			want: []string{"hello world"},
		},
		{
			name: "single variable",
			tpl:  "hello $1",
			want: []string{"hello ", "$1"},
		},
		{
			name: "multiple variables",
			tpl:  "$1 says hello to $2",
			want: []string{"$1", " says hello to ", "$2"},
		},
		{
			name: "variable at end",
			tpl:  "say hello $1",
			want: []string{"say hello ", "$1"},
		},
		{
			name: "empty string",
			tpl:  "",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.segment(tt.tpl)
			if len(got) != len(tt.want) {
				t.Errorf("segment() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("segment()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCore_matchTemplate(t *testing.T) {
	source := &mockSource{templates: node.Set{}}
	c := &core{
		source: source,
		varReg: variable.VarReg,
	}

	tests := []struct {
		name       string
		target     string
		template   string
		wantMatch  bool
		wantVarLen int
		wantVars   map[string]string // expected variable key -> value
	}{
		{
			name:       "exact match no vars",
			target:     "hello world",
			template:   "hello world",
			wantMatch:  true,
			wantVarLen: 0,
			wantVars:   map[string]string{},
		},
		{
			name:       "single variable",
			target:     "hello world",
			template:   "hello $1",
			wantMatch:  true,
			wantVarLen: 1,
			wantVars:   map[string]string{"$1": "world"},
		},
		{
			name:       "multiple variables",
			target:     "john says hello to mary",
			template:   "$1 says hello to $2",
			wantMatch:  true,
			wantVarLen: 2,
			wantVars:   map[string]string{"$1": "john", "$2": "mary"},
		},
		{
			name:       "variable at start",
			target:     "john says hello",
			template:   "$1 says hello",
			wantMatch:  true,
			wantVarLen: 1,
			wantVars:   map[string]string{"$1": "john"},
		},
		{
			name:       "variable at end",
			target:     "say hello to mary",
			template:   "say hello to $1",
			wantMatch:  true,
			wantVarLen: 1,
			wantVars:   map[string]string{"$1": "mary"},
		},
		{
			name:       "variable in middle",
			target:     "hello world there",
			template:   "hello $1 there",
			wantMatch:  true,
			wantVarLen: 1,
			wantVars:   map[string]string{"$1": "world"},
		},
		{
			name:       "no match different text",
			target:     "hello world",
			template:   "goodbye world",
			wantMatch:  false,
			wantVarLen: 0,
			wantVars:   nil,
		},
		{
			name:       "empty template",
			target:     "hello world",
			template:   "",
			wantMatch:  false,
			wantVarLen: 0,
			wantVars:   nil,
		},
		{
			name:       "target shorter than template",
			target:     "hi",
			template:   "hello world",
			wantMatch:  false,
			wantVarLen: 0,
			wantVars:   nil,
		},
		{
			name:       "single variable v2",
			target:     "Zero says [input_Cs] zero says if Zero says $1 to Susie to Susie to Susie",
			template:   "Zero says [input_Cs] $1 to Susie",
			wantMatch:  true,
			wantVarLen: 1,
			wantVars:   map[string]string{"$1": "zero says if Zero says $1 to Susie to Susie"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, vars := c.matchTemplate(tt.target, tt.template)
			if match != tt.wantMatch {
				t.Errorf("matchTemplate() match = %v, want %v", match, tt.wantMatch)
			}
			if len(vars) != tt.wantVarLen {
				t.Errorf("matchTemplate() vars len = %d, want %d", len(vars), tt.wantVarLen)
			}
			// Check variable values
			if tt.wantVars != nil {
				for key, wantVal := range tt.wantVars {
					if v, ok := vars[key]; !ok {
						t.Errorf("matchTemplate() missing variable %s", key)
					} else if v.Value() != wantVal {
						t.Errorf("matchTemplate() variable %s = %q, want %q", key, v.Value(), wantVal)
					}
				}
				// Check no extra variables
				for key := range vars {
					if _, ok := tt.wantVars[key]; !ok {
						t.Errorf("matchTemplate() unexpected variable %s = %q", key, vars[key].Value())
					}
				}
			}
		})
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		name    string
		varFrom variable.Set
		varTo   variable.Set
		wantErr bool
	}{
		{
			name:    "empty sets",
			varFrom: variable.Set{},
			varTo:   variable.Set{},
			wantErr: false,
		},
		{
			name: "copy value from varFrom to varTo",
			varFrom: variable.Set{
				"$1": variable.New("$1", "world"),
			},
			varTo: variable.Set{
				"$1": variable.New("$1", ""),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := clean(tt.varFrom, tt.varTo)
			if (err != nil) != tt.wantErr {
				t.Errorf("clean() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultCore(t *testing.T) {
	core := DefaultCore()
	if core == nil {
		t.Error("DefaultCore() returned nil")
	}
}
func TestDefaultSource(t *testing.T) {
	err := storage.DefaultJSONRepo().Init()
	if err != nil {
		t.Fatalf("DefaultSource() error = %v", err)
	}
	s := sourceInstance
	if s == nil {
		t.Fatal("DefaultSource() returned nil")
	}
	nodeSet, err := s.GetTemplates()
	if err != nil {
		t.Fatalf("DefaultSource() error = %v", err)
	}
	if len(nodeSet) == 0 {
		t.Fatal("DefaultSource() returned empty node set")
	}
	mustExisted := "Susie should say $1"
	if nodeSet[mustExisted] == nil {
		t.Fatal("DefaultSource() missing template")
	}
}
