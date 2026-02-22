package rules

import (
	"strings"
	"testing"

	"github.com/romano/envguard/internal/parser"
)

func TestInvalidChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wantN int
	}{
		{
			name:  "valid key",
			input: "FOO_BAR=value",
			wantN: 0,
		},
		{
			name:  "valid key with digits",
			input: "FOO_123=value",
			wantN: 0,
		},
		{
			name:  "key starting with underscore",
			input: "_FOO=value",
			wantN: 0,
		},
		{
			name:  "key starting with digit",
			input: "1FOO=value",
			wantN: 1,
		},
		{
			name:  "key with hyphen",
			input: "FOO-BAR=value",
			wantN: 1,
		},
		{
			name:  "key with dot",
			input: "FOO.BAR=value",
			wantN: 1,
		},
		{
			name:  "key with space",
			input: "FOO BAR=value",
			wantN: 1,
		},
		{
			name:  "lowercase key is valid",
			input: "foo_bar=value",
			wantN: 0,
		},
		{
			name:  "multiple invalid keys",
			input: "1FOO=a\nBAR-BAZ=b",
			wantN: 2,
		},
	}

	rule := &InvalidChars{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatal(err)
			}
			findings := rule.Run(file)
			if len(findings) != tt.wantN {
				t.Errorf("got %d findings, want %d", len(findings), tt.wantN)
				for _, f := range findings {
					t.Logf("  %s", f.Message)
				}
			}
			for _, f := range findings {
				if f.Severity != Error {
					t.Errorf("expected Error severity, got %s", f.Severity)
				}
			}
		})
	}
}
