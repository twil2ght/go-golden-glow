package template

import (
	"goldenglow/pkg/variable"
	"testing"
)

func TestMatchTemplate(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		template string
		wantOk   bool
		wantVars map[string]string
	}{
		{
			name:     "empty template",
			target:   "hello world",
			template: "",
			wantOk:   false,
			wantVars: nil,
		},
		{
			name:     "exact match no variables",
			target:   "hello",
			template: "hello",
			wantOk:   true,
			wantVars: map[string]string{},
		},
		{
			name:     "no match no variables",
			target:   "hello",
			template: "world",
			wantOk:   false,
			wantVars: nil,
		},
		{
			name:     "single variable match",
			target:   "hello world",
			template: "hello $1",
			wantOk:   true,
			wantVars: map[string]string{"$1": "world"},
		},
		{
			name:     "single variable no match",
			target:   "hello world",
			template: "goodbye $1",
			wantOk:   false,
			wantVars: nil,
		},
		{
			name:     "variable at start",
			target:   "hello world",
			template: "$1 world",
			wantOk:   true,
			wantVars: map[string]string{"$1": "hello"},
		},
		{
			name:     "variable in middle",
			target:   "the time is 14:30 now",
			template: "the time is $1 now",
			wantOk:   true,
			wantVars: map[string]string{"$1": "14:30"},
		},
		{
			name:     "multiple variables",
			target:   "Alice gives Bob a book",
			template: "$1 gives $2 a $3",
			wantOk:   true,
			wantVars: map[string]string{"$1": "Alice", "$2": "Bob", "$3": "book"},
		},
		{
			name:     "consecutive variables",
			target:   "hello world",
			template: "$1$2",
			wantOk:   true,
			wantVars: map[string]string{"$1": "h", "$2": "ello world"},
		},
		{
			name:     "only variables",
			target:   "hello world",
			template: "$1",
			wantOk:   true,
			wantVars: map[string]string{"$1": "hello world"},
		},
		{
			name:     "template longer than target",
			target:   "hi",
			template: "hello $1",
			wantOk:   false,
			wantVars: nil,
		},
		{
			name:     "special regex chars in template",
			target:   "price: $50.00",
			template: "price: $1",
			wantOk:   true,
			wantVars: map[string]string{"$1": "$50.00"},
		},
		{
			name:     "special regex chars in literal part are escaped",
			target:   "(hello) world",
			template: "$1 world",
			wantOk:   true,
			wantVars: map[string]string{"$1": "(hello)"},
		},
		{
			name:     "bracketed node tag in template",
			target:   "[node:executor] [value:hello]",
			template: "[node:executor] [value:$1]",
			wantOk:   true,
			wantVars: map[string]string{"$1": "hello"},
		},
		{
			name:     "partial match is not a match",
			target:   "hello world",
			template: "hello",
			wantOk:   false,
			wantVars: nil,
		},
		{
			name:     "partial match is not a match reverse",
			target:   "hello",
			template: "hello world",
			wantOk:   false,
			wantVars: nil,
		},
		{
			name:     "multi-digit variable",
			target:   "first second third",
			template: "$10 $20 $30",
			wantOk:   true,
			wantVars: map[string]string{"$10": "first", "$20": "second", "$30": "third"},
		},
		{
			name:     "variable at end captures rest",
			target:   "the quick brown fox",
			template: "the $1",
			wantOk:   true,
			wantVars: map[string]string{"$1": "quick brown fox"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, vars := MatchTemplate(tt.target, tt.template)

			if ok != tt.wantOk {
				t.Errorf("MatchTemplate() ok = %v, want %v", ok, tt.wantOk)
			}

			if tt.wantVars == nil {
				if vars != nil {
					t.Errorf("MatchTemplate() vars = %v, want nil", vars)
				}
				return
			}

			if len(vars) != len(tt.wantVars) {
				t.Errorf("MatchTemplate() got %d vars, want %d: got=%v", len(vars), len(tt.wantVars), varNames(vars))
				return
			}

			for k, wantVal := range tt.wantVars {
				item, ok := vars[k]
				if !ok {
					t.Errorf("MatchTemplate() missing key %q", k)
					continue
				}
				if item.Value() != wantVal {
					t.Errorf("MatchTemplate() vars[%q] = %q, want %q", k, item.Value(), wantVal)
				}
			}
		})
	}
}

func TestSegment(t *testing.T) {
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
			name: "one variable at end",
			tpl:  "hello $1",
			want: []string{"hello ", "$1"},
		},
		{
			name: "one variable at start",
			tpl:  "$1 world",
			want: []string{"$1", " world"},
		},
		{
			name: "one variable in middle",
			tpl:  "hello $1 world",
			want: []string{"hello ", "$1", " world"},
		},
		{
			name: "multiple variables",
			tpl:  "$1 gives $2 a $3",
			want: []string{"$1", " gives ", "$2", " a ", "$3"},
		},
		{
			name: "consecutive variables",
			tpl:  "$1$2",
			want: []string{"$1", "$2"},
		},
		{
			name: "only a variable",
			tpl:  "$1",
			want: []string{"$1"},
		},
		{
			name: "empty string",
			tpl:  "",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := segment(tt.tpl)
			if len(got) != len(tt.want) {
				t.Errorf("segment() = %v (%d elems), want %v (%d elems)", got, len(got), tt.want, len(tt.want))
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

func varNames(s variable.Set) []string {
	names := make([]string, 0, len(s))
	for k := range s {
		names = append(names, k)
	}
	return names
}
