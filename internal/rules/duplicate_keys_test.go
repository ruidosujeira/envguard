package rules

import (
	"strings"
	"testing"

	"github.com/romano/envguard/internal/parser"
)

func TestDuplicateKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantN    int
		wantMsgs []string
	}{
		{
			name:  "no duplicates",
			input: "FOO=1\nBAR=2\nBAZ=3",
			wantN: 0,
		},
		{
			name:  "one duplicate",
			input: "FOO=1\nBAR=2\nFOO=3",
			wantN: 1,
			wantMsgs: []string{"Duplicate key: FOO (first defined at line 1)"},
		},
		{
			name:  "multiple duplicates",
			input: "FOO=1\nBAR=2\nFOO=3\nBAR=4",
			wantN: 2,
		},
		{
			name:  "triple duplicate",
			input: "FOO=1\nFOO=2\nFOO=3",
			wantN: 2,
		},
		{
			name:  "comments and blanks ignored",
			input: "FOO=1\n# comment\n\nFOO=2",
			wantN: 1,
		},
	}

	rule := &DuplicateKeys{}
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
			for i, msg := range tt.wantMsgs {
				if i < len(findings) && findings[i].Message != msg {
					t.Errorf("finding[%d].Message = %q, want %q", i, findings[i].Message, msg)
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
