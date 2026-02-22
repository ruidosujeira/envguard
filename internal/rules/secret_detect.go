package rules

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/romano/envguard/internal/parser"
)

// SecretDetect identifies possible secrets exposed in .env files.
// Uses two strategies:
// 1. Pattern matching for known secret formats (AWS keys, JWTs, etc.)
// 2. Shannon entropy analysis for values that look like random secrets.
type SecretDetect struct{}

func (sd *SecretDetect) Name() string { return "secret-detect" }

// secretPattern defines a regex pattern and description for a known secret type.
type secretPattern struct {
	name    string
	pattern *regexp.Regexp
}

var secretPatterns = []secretPattern{
	{"AWS Access Key ID", regexp.MustCompile(`^AKIA[0-9A-Z]{16}$`)},
	{"AWS Secret Access Key", regexp.MustCompile(`^[A-Za-z0-9/+=]{40}$`)},
	{"JWT token", regexp.MustCompile(`^eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`)},
	{"RSA private key", regexp.MustCompile(`-----BEGIN (RSA |EC )?PRIVATE KEY-----`)},
	{"GitHub token", regexp.MustCompile(`^gh[ps]_[A-Za-z0-9_]{36,}$`)},
	{"GitHub token (classic)", regexp.MustCompile(`^ghp_[A-Za-z0-9_]{36,}$`)},
	{"Slack token", regexp.MustCompile(`^xox[baprs]-[A-Za-z0-9-]+$`)},
	{"Stripe secret key", regexp.MustCompile(`^sk_(live|test)_[A-Za-z0-9]{20,}$`)},
	{"Stripe publishable key", regexp.MustCompile(`^pk_(live|test)_[A-Za-z0-9]{20,}$`)},
	{"Heroku API key", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)},
	{"Generic password in URL", regexp.MustCompile(`://[^:]+:[^@]+@`)},
	{"SendGrid API key", regexp.MustCompile(`^SG\.[A-Za-z0-9_-]{22}\.[A-Za-z0-9_-]{43}$`)},
	{"Twilio API key", regexp.MustCompile(`^SK[0-9a-fA-F]{32}$`)},
}

// secretKeyHints are key names that typically hold secrets.
var secretKeyHints = []string{
	"SECRET",
	"PASSWORD",
	"PASSWD",
	"TOKEN",
	"PRIVATE_KEY",
	"API_KEY",
	"APIKEY",
	"AUTH",
	"CREDENTIAL",
}

// entropyThreshold is the minimum Shannon entropy to flag a value as a possible secret.
const entropyThreshold = 4.0

// minValueLength is the minimum length for entropy analysis.
const minValueLength = 8

func (sd *SecretDetect) Run(file *parser.EnvFile) []Finding {
	var findings []Finding

	for _, e := range file.Keys() {
		key := strings.TrimSpace(e.Key)
		value := e.Value

		if value == "" {
			continue
		}

		// Strategy 1: check value against known secret patterns.
		for _, sp := range secretPatterns {
			if sp.pattern.MatchString(value) {
				findings = append(findings, Finding{
					RuleName: sd.Name(),
					Severity: Error,
					Message:  fmt.Sprintf("Possible secret detected: %s (matches %s pattern)", key, sp.name),
					Line:     e.LineNumber,
					Column:   1,
				})
				break
			}
		}

		// Strategy 2: if the key name hints at a secret, check entropy.
		if isSecretKeyName(key) && len(value) >= minValueLength {
			entropy := shannonEntropy(value)
			if entropy >= entropyThreshold {
				// Avoid duplicate findings if pattern already matched.
				alreadyFound := false
				for _, f := range findings {
					if f.Line == e.LineNumber {
						alreadyFound = true
						break
					}
				}
				if !alreadyFound {
					findings = append(findings, Finding{
						RuleName: sd.Name(),
						Severity: Error,
						Message:  fmt.Sprintf("Possible secret detected: %s (high entropy value: %.1f)", key, entropy),
						Line:     e.LineNumber,
						Column:   1,
					})
				}
			}
		}
	}

	return findings
}

// isSecretKeyName checks if the key name suggests it holds a secret.
func isSecretKeyName(key string) bool {
	upper := strings.ToUpper(key)
	for _, hint := range secretKeyHints {
		if strings.Contains(upper, hint) {
			return true
		}
	}
	return false
}

// shannonEntropy calculates the Shannon entropy of a string.
// Higher entropy means more randomness, which suggests a secret value.
func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	freq := make(map[rune]int)
	for _, r := range s {
		freq[r]++
	}

	length := float64(len([]rune(s)))
	entropy := 0.0
	for _, count := range freq {
		p := float64(count) / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}
