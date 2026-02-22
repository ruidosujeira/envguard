package rules

import (
	"fmt"
	"strings"

	"github.com/romano/envguard/internal/parser"
)

// Formatting detects formatting issues:
// - Trailing whitespace
// - Spaces in key names (before the '=')
// - Excessive blank lines (more than 1 consecutive)
type Formatting struct{}

func (f *Formatting) Name() string { return "formatting" }

func (f *Formatting) Run(file *parser.EnvFile) []Finding {
	var findings []Finding

	consecutiveBlanks := 0

	for _, e := range file.Entries {
		// Trailing whitespace.
		if e.RawLine != strings.TrimRight(e.RawLine, " \t") {
			findings = append(findings, Finding{
				RuleName: f.Name(),
				Severity: Info,
				Message:  "Trailing whitespace",
				Line:     e.LineNumber,
				Column:   len(strings.TrimRight(e.RawLine, " \t")) + 1,
			})
		}

		// Excessive blank lines.
		if e.IsBlank {
			consecutiveBlanks++
			if consecutiveBlanks > 1 {
				findings = append(findings, Finding{
					RuleName: f.Name(),
					Severity: Info,
					Message:  "Excessive blank lines",
					Line:     e.LineNumber,
					Column:   1,
				})
			}
			continue
		}
		consecutiveBlanks = 0

		// Spaces in key names.
		if !e.IsComment && !e.IsBlank {
			key := e.Key
			trimmedKey := strings.TrimSpace(key)
			if key != trimmedKey {
				findings = append(findings, Finding{
					RuleName: f.Name(),
					Severity: Warning,
					Message:  fmt.Sprintf("Key %q has surrounding whitespace", trimmedKey),
					Line:     e.LineNumber,
					Column:   1,
				})
			}
		}
	}

	return findings
}
