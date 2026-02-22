package rules

import (
	"fmt"
	"strings"

	"github.com/romano/envguard/internal/parser"
)

// EmptyValues detects variables that have no value assigned.
type EmptyValues struct{}

func (e *EmptyValues) Name() string { return "empty-values" }

func (e *EmptyValues) Run(file *parser.EnvFile) []Finding {
	var findings []Finding

	for _, entry := range file.Keys() {
		// Only flag if there's an '=' sign but no value (and not a quoted empty string).
		if strings.Contains(entry.RawLine, "=") && entry.Value == "" && entry.Quoted == 0 {
			findings = append(findings, Finding{
				RuleName: e.Name(),
				Severity: Info,
				Message:  fmt.Sprintf("Empty value: %s", strings.TrimSpace(entry.Key)),
				Line:     entry.LineNumber,
				Column:   1,
			})
		}
	}

	return findings
}
