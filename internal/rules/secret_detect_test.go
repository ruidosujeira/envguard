package rules

import (
	"math"
	"strings"
	"testing"

	"github.com/romano/envguard/internal/parser"
)

func TestSecretDetect(t *testing.T) {
	tests := []struct {
		name  string
		input string
		wantN int
	}{
		{
			name:  "no secrets",
			input: "FOO=bar\nBAZ=hello",
			wantN: 0,
		},
		{
			name:  "AWS access key",
			input: "AWS_KEY=AKIAIOSFODNN7EXAMPLE",
			wantN: 1,
		},
		{
			name:  "JWT token",
			input: "TOKEN=eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.abc123def456",
			wantN: 1,
		},
		{
			name:  "Stripe secret key",
			input: "STRIPE_KEY=sk_test_abc123def456ghi789jkl012mno",
			wantN: 1,
		},
		{
			name:  "GitHub token",
			input: "GH_TOKEN=ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijkl",
			wantN: 1,
		},
		{
			name:  "password in URL",
			input: "DATABASE_URL=postgres://user:s3cr3tP@ss@host/db",
			wantN: 1,
		},
		{
			name:  "high entropy secret key",
			input: "SECRET_KEY=aB3$kL9!mN2@pQ5&rT8*",
			wantN: 1,
		},
		{
			name:  "low entropy password not flagged",
			input: "API_KEY=test",
			wantN: 0,
		},
		{
			name:  "empty value not flagged",
			input: "SECRET_KEY=",
			wantN: 0,
		},
		{
			name:  "private key header",
			input: "PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----",
			wantN: 1,
		},
	}

	rule := &SecretDetect{}
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

func TestShannonEntropy(t *testing.T) {
	tests := []struct {
		input   string
		wantMin float64
		wantMax float64
	}{
		{"", 0, 0},
		{"aaaa", 0, 0.1},
		{"abcd", 1.9, 2.1},
		{"aB3$kL9!mN2@pQ5&rT8*", 4.0, 5.0},
		{"password", 2.5, 3.5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := shannonEntropy(tt.input)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("shannonEntropy(%q) = %f, want between %f and %f", tt.input, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestIsSecretKeyName(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"SECRET_KEY", true},
		{"DATABASE_PASSWORD", true},
		{"API_KEY", true},
		{"AUTH_TOKEN", true},
		{"FOO_BAR", false},
		{"DATABASE_URL", false},
		{"PORT", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := isSecretKeyName(tt.key); got != tt.want {
				t.Errorf("isSecretKeyName(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestShannonEntropyEmpty(t *testing.T) {
	got := shannonEntropy("")
	if math.Abs(got) > 0.001 {
		t.Errorf("shannonEntropy(\"\") = %f, want 0", got)
	}
}
