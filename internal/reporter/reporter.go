package reporter

import (
	"github.com/romano/envguard/internal/rules"
	"github.com/romano/envguard/internal/sync"
)

// Report contains all findings and sync results for reporting.
type Report struct {
	Files []FileReport
}

// FileReport contains findings for a single file.
type FileReport struct {
	Path       string
	Findings   []rules.Finding
	SyncResult *sync.Result
}

// Counts returns the number of errors, warnings, and infos across all files.
func (r *Report) Counts() (errors, warnings, infos int) {
	for _, fr := range r.Files {
		for _, f := range fr.Findings {
			switch f.Severity {
			case rules.Error:
				errors++
			case rules.Warning:
				warnings++
			case rules.Info:
				infos++
			}
		}
	}
	return
}

// HasErrors returns true if any finding has Error severity.
func (r *Report) HasErrors() bool {
	e, _, _ := r.Counts()
	return e > 0
}

// HasFindings returns true if there are any findings at all.
func (r *Report) HasFindings() bool {
	for _, fr := range r.Files {
		if len(fr.Findings) > 0 {
			return true
		}
		if fr.SyncResult != nil && !fr.SyncResult.InSync() {
			return true
		}
	}
	return false
}

// Reporter formats and outputs a report.
type Reporter interface {
	Report(report *Report) (string, error)
}
