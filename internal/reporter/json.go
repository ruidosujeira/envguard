package reporter

import (
	"encoding/json"
)

// JSONReporter outputs findings as JSON.
type JSONReporter struct{}

type jsonOutput struct {
	Files   []jsonFileOutput `json:"files"`
	Summary jsonSummary      `json:"summary"`
}

type jsonFileOutput struct {
	Path     string        `json:"path"`
	Findings []jsonFinding `json:"findings"`
	Sync     *jsonSync     `json:"sync,omitempty"`
}

type jsonFinding struct {
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

type jsonSync struct {
	MissingInExample []string `json:"missing_in_example,omitempty"`
	MissingInEnv     []string `json:"missing_in_env,omitempty"`
}

type jsonSummary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Infos    int `json:"infos"`
}

func (jr *JSONReporter) Report(report *Report) (string, error) {
	output := jsonOutput{}

	for _, fr := range report.Files {
		jf := jsonFileOutput{
			Path:     fr.Path,
			Findings: make([]jsonFinding, 0, len(fr.Findings)),
		}

		for _, f := range fr.Findings {
			jf.Findings = append(jf.Findings, jsonFinding{
				Rule:     f.RuleName,
				Severity: f.Severity.String(),
				Message:  f.Message,
				Line:     f.Line,
				Column:   f.Column,
			})
		}

		if fr.SyncResult != nil && !fr.SyncResult.InSync() {
			jf.Sync = &jsonSync{
				MissingInExample: fr.SyncResult.MissingInExample,
				MissingInEnv:     fr.SyncResult.MissingInEnv,
			}
		}

		output.Files = append(output.Files, jf)
	}

	errors, warnings, infos := report.Counts()
	output.Summary = jsonSummary{
		Errors:   errors,
		Warnings: warnings,
		Infos:    infos,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data) + "\n", nil
}
