package rules

import (
	"strings"
	"testing"

	"github.com/romano/envguard/internal/parser"
)

func TestEmptyValues(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wantN int
	}{
		{
			name:  "no empty values",
			input: "FOO=bar\nBAZ=qux",
			wantN: 0,
		},
		{
			name:  "empty value with equals",
			input: "FOO=\nBAR=value",
			wantN: 1,
		},
		{
			name:  "multiple empty values",
			input: "FOO=\nBAR=\nBAZ=value",
			wantN: 2,
		},
		{
			name:  "key without equals is not flagged",
			input: "FOO",
			wantN: 0,
		},
		{
			name:  "quoted empty string is not empty",
			input: `FOO=""`,
			wantN: 0,
		},
		{
			name:  "comments ignored",
			input: "# FOO=\nBAR=value",
			wantN: 0,
		},
	}

	rule := &EmptyValues{}
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
				if f.Severity != Info {
					t.Errorf("expected Info severity, got %s", f.Severity)
				}
			}
		})
	}
}
