package rules

import (
	"github.com/romano/envguard/internal/parser"
)

// Severity indicates how serious a finding is.
type Severity int

const (
	Info    Severity = iota // suggestion (formatting, ordering)
	Warning                 // should fix (typos, missing .gitignore)
	Error                   // must fix (duplicates, exposed secrets)
)

func (s Severity) String() string {
	switch s {
	case Info:
		return "info"
	case Warning:
		return "warning"
	case Error:
		return "error"
	default:
		return "unknown"
	}
}

// Finding represents a single issue found by a rule.
type Finding struct {
	RuleName string
	Severity Severity
	Message  string
	Line     int
	Column   int
}

// Rule analyzes an EnvFile and returns findings.
type Rule interface {
	Name() string
	Run(file *parser.EnvFile) []Finding
}

// RunAll executes all given rules against the file and returns all findings.
func RunAll(file *parser.EnvFile, rules []Rule) []Finding {
	var all []Finding
	for _, r := range rules {
		all = append(all, r.Run(file)...)
	}
	return all
}

// DefaultRules returns the standard set of lint rules.
func DefaultRules() []Rule {
	return []Rule{
		&DuplicateKeys{},
		&EmptyValues{},
		&InvalidChars{},
		&Formatting{},
		&TypoDetect{},
	}
}
