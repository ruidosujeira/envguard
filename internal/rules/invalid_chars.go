package rules

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/romano/envguard/internal/parser"
)

// InvalidChars detects keys with characters that are not valid for environment variable names.
// Valid characters are: uppercase/lowercase letters, digits, and underscores.
// Keys should start with a letter or underscore (not a digit).
type InvalidChars struct{}

func (ic *InvalidChars) Name() string { return "invalid-chars" }

func (ic *InvalidChars) Run(file *parser.EnvFile) []Finding {
	var findings []Finding

	for _, e := range file.Keys() {
		key := strings.TrimSpace(e.Key)
		if key == "" {
			continue
		}

		// Check first character: must be letter or underscore.
		first := rune(key[0])
		if !unicode.IsLetter(first) && first != '_' {
			findings = append(findings, Finding{
				RuleName: ic.Name(),
				Severity: Error,
				Message:  fmt.Sprintf("Invalid key %q: must start with a letter or underscore", key),
				Line:     e.LineNumber,
				Column:   1,
			})
			continue
		}

		// Check remaining characters.
		for i, r := range key {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				findings = append(findings, Finding{
					RuleName: ic.Name(),
					Severity: Error,
					Message:  fmt.Sprintf("Invalid character %q in key %q at position %d", string(r), key, i+1),
					Line:     e.LineNumber,
					Column:   i + 1,
				})
				break
			}
		}
	}

	return findings
}
