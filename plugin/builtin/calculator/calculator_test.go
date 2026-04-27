package calculator

import (
	"testing"
)

func TestCalculate_Fixes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Bug 1: operators inside parentheses - fixed
		{name: "parens with external mult", input: "(1+2)*3", expected: "9"},
		{name: "parens with external mult reversed", input: "3*(1+2)", expected: "9"},
		{name: "two paren groups added", input: "(1+2)+(3+4)", expected: "10"},
		{name: "two paren groups multiplied", input: "(1+2)*(3+4)", expected: "21"},
		{name: "nested expression", input: "2*(3+4)", expected: "14"},
		{name: "complex parens", input: "(10-3)*(2+1)", expected: "21"},
		{name: "operator inside single paren group", input: "1+(2*3)", expected: "7"},
		{name: "subtraction outside parens", input: "(5+5)-(2+3)", expected: "5"},

		// Bug 3: unary minus before parens - fixed
		{name: "unary minus before parens", input: "-(3+5)", expected: "-8"},
		{name: "unary minus simple", input: "-5", expected: "-5"},
		{name: "unary minus with addition", input: "-3+5", expected: "2"},
		{name: "double unary minus", input: "--5", expected: "5"},
		{name: "unary minus before parens with mult", input: "-(2*3)", expected: "-6"},

		// Existing functionality preserved
		{name: "simple add", input: "1+2", expected: "3"},
		{name: "simple sub", input: "5-3", expected: "2"},
		{name: "simple mult", input: "4*3", expected: "12"},
		{name: "simple div", input: "10/2", expected: "5"},
		{name: "precedence add mul", input: "1+2*3", expected: "7"},
		{name: "precedence mul add", input: "1*2+3", expected: "5"},
		{name: "nested parens", input: "((1+2))", expected: "3"},
		{name: "div by 2 twice", input: "8/2/2", expected: "2"},
		{name: "float result", input: "5/2", expected: "2.5"},

		// Bug 2: naive paren check - fixed (e.g., (1)+(2) no longer breaks)
		{name: "consecutive paren nums added", input: "(1)+(2)", expected: "3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewCalculator().(*calculator)
			result, err := calc.calculate(tt.input)
			if err != nil {
				t.Errorf("calculate(%q) unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("calculate(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
