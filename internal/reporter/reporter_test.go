package reporter

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/romano/envguard/internal/rules"
	"github.com/romano/envguard/internal/sync"
)

func TestTextReporter_NoFindings(t *testing.T) {
	r := &TextReporter{Color: false}
	report := &Report{}
	out, err := r.Report(report)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "No issues found") {
		t.Errorf("expected 'No issues found', got: %s", out)
	}
}

func TestTextReporter_WithFindings(t *testing.T) {
	r := &TextReporter{Color: false}
	report := &Report{
		Files: []FileReport{
			{
				Path: ".env",
				Findings: []rules.Finding{
					{RuleName: "duplicate-keys", Severity: rules.Error, Message: "Duplicate key: FOO", Line: 3, Column: 1},
					{RuleName: "typo-detect", Severity: rules.Warning, Message: "Possible typo: DATBASE_URL", Line: 5, Column: 1},
					{RuleName: "empty-values", Severity: rules.Info, Message: "Empty value: REDIS_URL", Line: 9, Column: 1},
				},
			},
		},
	}

	out, err := r.Report(report)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, ".env") {
		t.Error("expected file path in output")
	}
	if !strings.Contains(out, "Duplicate key: FOO") {
		t.Error("expected finding message in output")
	}
	if !strings.Contains(out, "1 error") {
		t.Error("expected error count in output")
	}
	if !strings.Contains(out, "1 warning") {
		t.Error("expected warning count in output")
	}
}

func TestTextReporter_WithSync(t *testing.T) {
	r := &TextReporter{Color: false}
	report := &Report{
		Files: []FileReport{
			{
				Path: ".env.example",
				SyncResult: &sync.Result{
					MissingInExample: []string{"NEW_KEY", "ANOTHER_KEY"},
				},
			},
		},
	}

	out, err := r.Report(report)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "NEW_KEY") {
		t.Error("expected missing key in output")
	}
	if !strings.Contains(out, "ANOTHER_KEY") {
		t.Error("expected missing key in output")
	}
}

func TestJSONReporter(t *testing.T) {
	r := &JSONReporter{}
	report := &Report{
		Files: []FileReport{
			{
				Path: ".env",
				Findings: []rules.Finding{
					{RuleName: "duplicate-keys", Severity: rules.Error, Message: "Duplicate key: FOO", Line: 3, Column: 1},
				},
			},
		},
	}

	out, err := r.Report(report)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	summary, ok := parsed["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("expected summary in JSON output")
	}
	if summary["errors"].(float64) != 1 {
		t.Errorf("expected 1 error, got %v", summary["errors"])
	}
}

func TestReport_Counts(t *testing.T) {
	report := &Report{
		Files: []FileReport{
			{
				Findings: []rules.Finding{
					{Severity: rules.Error},
					{Severity: rules.Error},
					{Severity: rules.Warning},
					{Severity: rules.Info},
				},
			},
			{
				Findings: []rules.Finding{
					{Severity: rules.Warning},
				},
			},
		},
	}

	errors, warnings, infos := report.Counts()
	if errors != 2 {
		t.Errorf("errors = %d, want 2", errors)
	}
	if warnings != 2 {
		t.Errorf("warnings = %d, want 2", warnings)
	}
	if infos != 1 {
		t.Errorf("infos = %d, want 1", infos)
	}
}

func TestReport_HasErrors(t *testing.T) {
	noErrors := &Report{
		Files: []FileReport{
			{Findings: []rules.Finding{{Severity: rules.Warning}}},
		},
	}
	if noErrors.HasErrors() {
		t.Error("expected no errors")
	}

	withErrors := &Report{
		Files: []FileReport{
			{Findings: []rules.Finding{{Severity: rules.Error}}},
		},
	}
	if !withErrors.HasErrors() {
		t.Error("expected errors")
	}
}

func TestPluralS(t *testing.T) {
	if pluralS(0) != "s" {
		t.Error("0 should be plural")
	}
	if pluralS(1) != "" {
		t.Error("1 should be singular")
	}
	if pluralS(2) != "s" {
		t.Error("2 should be plural")
	}
}
