package expression

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestFunctionCallExpression(t *testing.T) {
	suffix := `{ <div>`
	tests := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{
			name:     "functions in other packages can be called",
			input:    `components.Other()` + suffix,
			expected: `components.Other()`,
		},
		{
			name:     "array indices are supported",
			input:    `components[0].Other()` + suffix,
			expected: `components[0].Other()`,
		},
		{
			name:     "map keys are supported",
			input:    `components["name"].Other()` + suffix,
			expected: `components["name"].Other()`,
		},
		{
			name:     "function literals as inputs are supported",
			input:    `components["name"].Other(func() bool { return true })` + suffix,
			expected: `components["name"].Other(func() bool { return true })`,
		},
		{
			name: "multi-line function calls are supported",
			input: `component(map[string]string{
				"namea": "name_a",
			  "nameb": "name_b",
			})` + suffix,
			expected: `component(map[string]string{
				"namea": "name_a",
			  "nameb": "name_b",
			})`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseExpression(test.input)
			if test.expectedErr == nil && err != nil {
				t.Fatalf("expected nil error, got %v, %T", err, err)
			}
			if test.expectedErr != nil && err == nil {
				t.Fatalf("expected err %q, got %v", test.expectedErr.Error(), err)
			}
			if test.expectedErr != nil && err != nil && test.expectedErr.Error() != err.Error() {
				t.Fatalf("expected err %q, got %q", test.expectedErr.Error(), err.Error())
			}
			if diff := cmp.Diff(test.expected, actual, cmpopts.EquateErrors()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
