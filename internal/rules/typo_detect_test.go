package rules

import (
	"strings"
	"testing"

	"github.com/romano/envguard/internal/parser"
)

func TestTypoDetect(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantN    int
		wantMsgs []string
	}{
		{
			name:  "exact match no flag",
			input: "DATABASE_URL=postgres://localhost/db",
			wantN: 0,
		},
		{
			name:  "typo of DATABASE_URL",
			input: "DATBASE_URL=postgres://localhost/db",
			wantN: 1,
			wantMsgs: []string{"Possible typo: DATBASE_URL (did you mean DATABASE_URL?)"},
		},
		{
			name:  "typo of API_KEY",
			input: "API_KYE=secret",
			wantN: 1,
		},
		{
			name:  "completely different key no flag",
			input: "MY_CUSTOM_VARIABLE=value",
			wantN: 0,
		},
		{
			name:  "near-match flags unknown key close to known key",
			input: "DATABASE_URL=foo\nDATABASE_URI=bar",
			wantN: 1, // DATABASE_URI is 1 edit from DATABASE_URL (a known key)
		},
		{
			name:  "unrelated custom keys not flagged",
			input: "MY_APP_TIMEOUT=30\nFEATURE_FLAG_X=true",
			wantN: 0,
		},
	}

	rule := &TypoDetect{}
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
		})
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"kitten", "sitting", 3},
		{"DATABASE_URL", "DATBASE_URL", 1},
		{"API_KEY", "API_KYE", 2},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			got := levenshtein(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
