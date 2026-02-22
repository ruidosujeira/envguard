package rules

import (
	"fmt"

	"github.com/romano/envguard/internal/parser"
)

// DuplicateKeys detects keys that appear more than once.
// The second occurrence silently overwrites the first, which is almost always a bug.
type DuplicateKeys struct{}

func (d *DuplicateKeys) Name() string { return "duplicate-keys" }

func (d *DuplicateKeys) Run(file *parser.EnvFile) []Finding {
	seen := make(map[string]int) // key -> first line number
	var findings []Finding

	for _, e := range file.Keys() {
		if firstLine, exists := seen[e.Key]; exists {
			findings = append(findings, Finding{
				RuleName: d.Name(),
				Severity: Error,
				Message:  fmt.Sprintf("Duplicate key: %s (first defined at line %d)", e.Key, firstLine),
				Line:     e.LineNumber,
				Column:   1,
			})
		} else {
			seen[e.Key] = e.LineNumber
		}
	}

	return findings
}
