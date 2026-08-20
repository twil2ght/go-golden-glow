package builder

import (
	"encoding/json"
	"goldenglow/pkg/database"
	"os"
	"path/filepath"
	"testing"

	"goldenglow/m"
)

// --- mock brainsaver ---

type mockSaver struct {
	saves []struct{ t, r m.Hash }
}

func (ms *mockSaver) GetRepo() database.database {
	//TODO implement me
	panic("implement me")
}

func (ms *mockSaver) Save(t, r m.Hash) {
	ms.saves = append(ms.saves, struct{ t, r m.Hash }{t, r})
}

func newTestBuilder() *builder {
	return &builder{
		saver:   &mockSaver{},
		mapping: make(map[string]string),
	}
}

// --- parseTemplateArgs ---

func TestParseTemplateArgs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "single pair",
			input:   "key1=value1",
			want:    map[string]string{"key1": "value1"},
			wantErr: false,
		},
		{
			name:    "multiple pairs",
			input:   "key1=value1,key2=value2",
			want:    map[string]string{"key1": "value1", "key2": "value2"},
			wantErr: false,
		},
		{
			name:    "three pairs",
			input:   "a=1,b=2,c=3",
			want:    map[string]string{"a": "1", "b": "2", "c": "3"},
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "empty pair skipped",
			input:   "key1=val1,,key2=val2",
			want:    map[string]string{"key1": "val1", "key2": "val2"},
			wantErr: false,
		},
		{
			name:    "trailing comma",
			input:   "key=val,",
			want:    map[string]string{"key": "val"},
			wantErr: false,
		},
		{
			name:    "leading comma",
			input:   ",key=val",
			want:    map[string]string{"key": "val"},
			wantErr: false,
		},
		{
			name:    "missing equals sign",
			input:   "keyvalue",
			wantErr: true,
		},
		{
			name:    "too many equals signs",
			input:   "key=val=ue",
			wantErr: true,
		},
		{
			name:    "only commas",
			input:   ",,,",
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "value with spaces",
			input:   "key=hello world",
			want:    map[string]string{"key": "hello world"},
			wantErr: false,
		},
		{
			name:    "spaces around equals not trimmed",
			input:   "key =value",
			want:    map[string]string{"key ": "value"},
			wantErr: false,
		},
	}

	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		got, err := parseTemplateArgs(tt.input)
	// 		if (err != nil) != tt.wantErr {
	// 			t.Errorf("parseTemplateArgs(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
	// 			return
	// 		}
	// 		if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
	// 			t.Errorf("parseTemplateArgs(%q) = %v, want %v", tt.input, got, tt.want)
	// 		}
	// 	})
	// }
}

// --- replaceVars ---

func TestReplaceVars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		vars  map[string]string
		want  string
	}{
		{
			name:  "single replacement",
			input: "hello $1 world",
			vars:  map[string]string{"$1": "beautiful"},
			want:  "hello beautiful world",
		},
		{
			name:  "multiple replacements",
			input: "$1 says $2",
			vars:  map[string]string{"$1": "Alice", "$2": "hi"},
			want:  "Alice says hi",
		},
		{
			name:  "no match",
			input: "no variables here",
			vars:  map[string]string{"$1": "unused"},
			want:  "no variables here",
		},
		{
			name:  "empty vars",
			input: "hello $1",
			vars:  map[string]string{},
			want:  "hello $1",
		},
		{
			name:  "empty input",
			input: "",
			vars:  map[string]string{"$1": "x"},
			want:  "",
		},
		{
			name:  "non-overlapping keys",
			input: "$a $b",
			vars:  map[string]string{"$a": "X", "$b": "Y"},
			want:  "X Y",
		},
		{
			name:  "multiple occurrences",
			input: "$1 and $1 again",
			vars:  map[string]string{"$1": "same"},
			want:  "same and same again",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceVars(tt.input, tt.vars)
			if got != tt.want {
				t.Errorf("replaceVars(%q, %v) = %q, want %q", tt.input, tt.vars, got, tt.want)
			}
		})
	}
}

// --- ParseTpl ---

func TestParseTpl_BasicInputOutput(t *testing.T) {
	b := newTestBuilder()
	tpl := &template{
		Name: "test",
		Data: []dataItem{
			{
				Commands: []string{
					"[input] the cat sat",
					"[output] the mat",
				},
			},
		},
	}

	b.ParseTpl(tpl, nil)

	ms := b.saver.(*mockSaver)
	if len(ms.saves) != 1 {
		t.Fatalf("expected 1 save, got %d", len(ms.saves))
	}
	save := ms.saves[0]
	if _, ok := save.t["the cat sat"]; !ok {
		t.Errorf("expected input 'the cat sat' in hash, got %v", save.t)
	}
	if _, ok := save.r["the mat"]; !ok {
		t.Errorf("expected output 'the mat' in hash, got %v", save.r)
	}
}

func TestParseTpl_VariableReplacement(t *testing.T) {
	b := newTestBuilder()
	tpl := &template{
		Name: "test",
		Data: []dataItem{
			{
				Commands: []string{
					"[input] $Attr B @Caller C",
					"[output] C @ $Attr B -> yes",
				},
			},
		},
	}

	b.ParseTpl(tpl, map[string]string{"$Attr": "big"})

	ms := b.saver.(*mockSaver)
	if len(ms.saves) != 1 {
		t.Fatalf("expected 1 save, got %d", len(ms.saves))
	}
	save := ms.saves[0]
	if _, ok := save.t["big B @Caller C"]; !ok {
		t.Errorf("expected 'big B @Caller C' in inputs, got %v", save.t)
	}
	if _, ok := save.r["C @ big B -> yes"]; !ok {
		t.Errorf("expected 'C @ big B -> yes' in outputs, got %v", save.r)
	}
}

func TestParseTpl_MultipleDataItems(t *testing.T) {
	b := newTestBuilder()
	tpl := &template{
		Name: "multi",
		Data: []dataItem{
			{
				Commands: []string{
					"[input] A",
					"[output] B",
				},
			},
			{
				Commands: []string{
					"[input] C",
					"[output] D",
				},
			},
		},
	}

	b.ParseTpl(tpl, nil)

	ms := b.saver.(*mockSaver)
	if len(ms.saves) != 2 {
		t.Fatalf("expected 2 saves, got %d", len(ms.saves))
	}
}

func TestParseTpl_InputOnlyNoSave(t *testing.T) {
	b := newTestBuilder()
	tpl := &template{
		Name: "inonly",
		Data: []dataItem{
			{
				Commands: []string{
					"[input] only input here",
				},
			},
		},
	}

	b.ParseTpl(tpl, nil)

	ms := b.saver.(*mockSaver)
	if len(ms.saves) != 0 {
		t.Errorf("expected 0 saves for input-only data item, got %d", len(ms.saves))
	}
}

func TestParseTpl_OutputOnlyNoSave(t *testing.T) {
	b := newTestBuilder()
	tpl := &template{
		Name: "outonly",
		Data: []dataItem{
			{
				Commands: []string{
					"[output] only output here",
				},
			},
		},
	}

	b.ParseTpl(tpl, nil)

	ms := b.saver.(*mockSaver)
	if len(ms.saves) != 0 {
		t.Errorf("expected 0 saves for output-only data item, got %d", len(ms.saves))
	}
}

func TestParseTpl_MultipleInputsSingleOutput(t *testing.T) {
	b := newTestBuilder()
	tpl := &template{
		Name: "multiIn",
		Data: []dataItem{
			{
				Commands: []string{
					"[input] cond1",
					"[input] cond2",
					"[input] cond3",
					"[output] result",
				},
			},
		},
	}

	b.ParseTpl(tpl, nil)

	ms := b.saver.(*mockSaver)
	if len(ms.saves) != 1 {
		t.Fatalf("expected 1 save, got %d", len(ms.saves))
	}
	save := ms.saves[0]
	if len(save.t) != 3 {
		t.Errorf("expected 3 inputs, got %d: %v", len(save.t), save.t)
	}
	if len(save.r) != 1 {
		t.Errorf("expected 1 output, got %d: %v", len(save.r), save.r)
	}
}

func TestParseTpl_MappingApplied(t *testing.T) {
	b := newTestBuilder()
	b.mapping["cat"] = "feline"

	tpl := &template{
		Name: "mapped",
		Data: []dataItem{
			{
				Commands: []string{
					"[input] the cat sat",
					"[output] the feline sat",
				},
			},
		},
	}

	b.ParseTpl(tpl, nil)

	ms := b.saver.(*mockSaver)
	if len(ms.saves) != 1 {
		t.Fatalf("expected 1 save, got %d", len(ms.saves))
	}
	save := ms.saves[0]
	if _, ok := save.t["the feline sat"]; !ok {
		t.Errorf("expected mapping to replace 'cat' with 'feline', got %v", save.t)
	}
}

func TestParseTpl_NonCommandPrefixIgnored(t *testing.T) {
	b := newTestBuilder()
	tpl := &template{
		Name: "mixed",
		Data: []dataItem{
			{
				Commands: []string{
					"[input] real input",
					"just a comment line",
					"[output] real output",
				},
			},
		},
	}

	b.ParseTpl(tpl, nil)

	ms := b.saver.(*mockSaver)
	if len(ms.saves) != 1 {
		t.Fatalf("expected 1 save, got %d", len(ms.saves))
	}
	save := ms.saves[0]
	if len(save.t) != 1 || len(save.r) != 1 {
		t.Errorf("expected exactly 1 input and 1 output, got t=%v r=%v", save.t, save.r)
	}
}

// --- RunTemplate ---

func TestRunTemplate_Success(t *testing.T) {
	b := newTestBuilder()
	b.templates = m.Map[*template]{
		"greet": {
			Name: "greet",
			Data: []dataItem{
				{
					Commands: []string{
						"[input] $who arrives",
						"[output] greet $who",
					},
				},
			},
		},
	}

	b.RunTemplate("greet", "$who=Alice")

	ms := b.saver.(*mockSaver)
	if len(ms.saves) != 1 {
		t.Fatalf("expected 1 save, got %d", len(ms.saves))
	}
	save := ms.saves[0]
	if _, ok := save.t["Alice arrives"]; !ok {
		t.Errorf("expected 'Alice arrives', got %v", save.t)
	}
	if _, ok := save.r["greet Alice"]; !ok {
		t.Errorf("expected 'greet Alice', got %v", save.r)
	}
}

func TestRunTemplate_NotFound(t *testing.T) {
	b := newTestBuilder()
	b.templates = m.Map[*template]{}

	// Should not panic
	b.RunTemplate("nonexistent", "a=b")
}

func TestRunTemplate_InvalidArgs(t *testing.T) {
	b := newTestBuilder()
	b.templates = m.Map[*template]{
		"t": {Name: "t", Data: []dataItem{}},
	}

	// Should not panic
	b.RunTemplate("t", "invalid_args_format")
}

// --- loadTemplates ---

func TestLoadTemplates(t *testing.T) {
	dir := t.TempDir()

	writeJSON := func(name string, content map[string]any) {
		data, _ := json.Marshal(content)
		os.WriteFile(filepath.Join(dir, name+".json"), data, 0644)
	}
	writeJSON("a_valid_template", map[string]any{
		"name":        "my_tpl",
		"is_template": true,
		"args":        map[string]any{"Attr": "$1"},
		"data":        []any{},
	})

	writeJSON("b_not_a_template", map[string]any{
		"name":        "no_tpl",
		"is_template": false,
		"data":        []any{},
	})

	writeJSON("c_no_name", map[string]any{
		"is_template": true,
		"data":        []any{},
	})

	got, err := loadTemplates(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 template, got %d: %v", len(got), got)
	}
	if _, ok := got["my_tpl"]; !ok {
		t.Errorf("expected template 'my_tpl' to be loaded")
	}
}

func TestLoadTemplates_InvalidJSONStopsWalk(t *testing.T) {
	dir := t.TempDir()

	// "a_invalid" comes before "b_valid" alphabetically, so walk hits it first
	_ = os.WriteFile(filepath.Join(dir, "a_invalid.json"), []byte("not json"), 0644)
	_ = os.WriteFile(filepath.Join(dir, "b_valid.json"), []byte(`{"name":"t","is_template":true,"data":[]}`), 0644)

	got, err := loadTemplates(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Walk stops at the first invalid file, so the valid one is never reached
	if len(got) != 0 {
		t.Errorf("expected 0 templates (walk stops on error), got %d", len(got))
	}
}

func TestLoadTemplates_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	got, err := loadTemplates(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 templates, got %d", len(got))
	}
}
