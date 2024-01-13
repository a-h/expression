package expression

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// If, Else, For, Range, Switch, Case, Default, Expression, Children

func TestSwitch(t *testing.T) {
	suffixes := []string{
		"\ncase 1:\n\t<div>\n\tcase 2:\n\t\t<div>\n\tdefault:\n\t\t<div>\n\t</div>}",
		"\ndefault:\n\t<div>\n\t</div>}",
		"\n}",
	}
	tests := []testInput{
		{
			name:  "switch",
			input: `switch {`,
		},
		{
			name:  "switch with expression",
			input: `switch x {`,
		},
		{
			name:  "switch with function call",
			input: `switch pkg.Func() {`,
		},
		{
			name:  "type switch",
			input: `switch x := x.(type) {`,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, suffix))
		}
	}
}

func TestCase(t *testing.T) {
	suffixes := []string{
		"\n<div>\ncase 1 content\n\t</div>\n\tcase 3:",
		"\ndefault:\n\t<div>\n\t</div>}",
		"\n}",
	}
	tests := []testInput{
		{
			name:  "case",
			input: `case 1:`,
		},
		{
			name:  "case with expression",
			input: `case x > 3:`,
		},
		{
			name:  "case with function call",
			input: `case pkg.Func():`,
		},
		{
			name:  "case with multiple expressions",
			input: `case x > 3, x < 4:`,
		},
		{
			name:  "case with multiple expressions and default",
			input: `case x > 3, x < 4, x == 5:`,
		},
		{
			name:  "case with type switch",
			input: `case bool:`,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, suffix))
		}
	}
}

func TestIf(t *testing.T) {
	suffixes := []string{
		"\n<div>\nif true content\n\t</div>}",
	}
	tests := []testInput{
		{
			name:  "basic if",
			input: `if true {`,
		},
		{
			name:  "if function call",
			input: `if pkg.Func() {`,
		},
		{
			name:  "compound",
			input: "if x := val(); x > 3 {",
		},
		{
			name:  "if multiple",
			input: `if x && y && (!z) {`,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, suffix))
		}
	}
}

func TestElse(t *testing.T) {
	suffixes := []string{
		"\n<div>\nelse content\n\t</div>}",
	}
	tests := []testInput{
		{
			name:  "else",
			input: `else {`,
		},
		{
			name:  "boolean",
			input: `else if true {`,
		},
		{
			name:  "func",
			input: `else if pkg.Func() {`,
		},
		{
			name:  "expression",
			input: "else if x > 3 {",
		},
		{
			name:  "multiple",
			input: `else if x && y && (!z) {`,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, suffix))
		}
	}
}

func TestFor(t *testing.T) {
	suffixes := []string{
		"\n<div>\nloop content\n\t</div>}",
	}
	tests := []testInput{
		{
			name:  "three component",
			input: `for i := 0; i < 100; i++ {`,
		},
		{
			name:  "three component, empty",
			input: `for ; ; i++ {`,
		},
		{
			name:  "while",
			input: `for n < 5 {`,
		},
		{
			name:  "infinite",
			input: `for {`,
		},
		{
			name:  "range with index",
			input: `for k, v := range m {`,
		},
		{
			name:  "range with key only",
			input: `for k := range m {`,
		},
		{
			name:  "channel receive",
			input: `for x := range channel {`,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, suffix))
		}
	}
}

func TestExpression(t *testing.T) {
	suffixes := []string{
		"}",
	}
	tests := []testInput{
		{
			name:  "function call in package",
			input: `components.Other()`,
		},
		{
			name:  "slice index call",
			input: `components[0].Other()`,
		},
		{
			name:  "map index function call",
			input: `components["name"].Other()`,
		},
		{
			name:  "function literal",
			input: `components["name"].Other(func() bool { return true })`,
		},
		{
			name: "multiline function call",
			input: `component(map[string]string{
				"namea": "name_a",
			  "nameb": "name_b",
			})`,
		},
	}
	for _, test := range tests {
		for i, suffix := range suffixes {
			t.Run(fmt.Sprintf("%s_%d", test.name, i), run(test, suffix))
		}
	}
}

type testInput struct {
	name        string
	input       string
	expectedErr error
}

func run(test testInput, suffix string) func(t *testing.T) {
	return func(t *testing.T) {
		actual, err := ParseExpression(test.input + suffix)
		if test.expectedErr == nil && err != nil {
			t.Fatalf("expected nil error, got %v, %T", err, err)
		}
		if test.expectedErr != nil && err == nil {
			t.Fatalf("expected err %q, got %v", test.expectedErr.Error(), err)
		}
		if test.expectedErr != nil && err != nil && test.expectedErr.Error() != err.Error() {
			t.Fatalf("expected err %q, got %q", test.expectedErr.Error(), err.Error())
		}
		if diff := cmp.Diff(test.input, actual, cmpopts.EquateErrors()); diff != "" {
			t.Error(diff)
		}
	}
}
