package rules

import (
	"strings"
	"testing"

	"github.com/romano/envguard/internal/parser"
)

func TestFormatting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantN    int
		wantMsgs []string
	}{
		{
			name:  "clean file",
			input: "FOO=bar\nBAZ=qux",
			wantN: 0,
		},
		{
			name:  "trailing whitespace",
			input: "FOO=bar   ",
			wantN: 1,
			wantMsgs: []string{"Trailing whitespace"},
		},
		{
			name:  "trailing tab",
			input: "FOO=bar\t",
			wantN: 1,
			wantMsgs: []string{"Trailing whitespace"},
		},
		{
			name:  "key with leading space",
			input: " FOO=bar",
			wantN: 1,
			wantMsgs: []string{`Key "FOO" has surrounding whitespace`},
		},
		{
			name:  "key with trailing space before equals",
			input: "FOO =bar",
			wantN: 1,
			wantMsgs: []string{`Key "FOO" has surrounding whitespace`},
		},
		{
			name:  "excessive blank lines",
			input: "FOO=1\n\n\nBAR=2",
			wantN: 1,
			wantMsgs: []string{"Excessive blank lines"},
		},
		{
			name:  "single blank line is ok",
			input: "FOO=1\n\nBAR=2",
			wantN: 0,
		},
		{
			name:  "multiple issues",
			input: " FOO =bar   ",
			wantN: 2, // trailing whitespace + key whitespace
		},
	}

	rule := &Formatting{}
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
					t.Logf("  line %d: %s", f.Line, f.Message)
				}
			}
			for i, msg := range tt.wantMsgs {
				if i < len(findings) && findings[i].Message != msg {
					t.Errorf("finding[%d].Message = %q, want %q", i, findings[i].Message, msg)
				}
			}
		})
	}
}
