package reporter

import (
	"fmt"
	"strings"

	"github.com/romano/envguard/internal/rules"
)

// TextReporter outputs human-readable text with optional colors.
type TextReporter struct {
	Color bool
}

func (tr *TextReporter) Report(report *Report) (string, error) {
	var b strings.Builder

	for _, fr := range report.Files {
		if len(fr.Findings) == 0 && (fr.SyncResult == nil || fr.SyncResult.InSync()) {
			continue
		}

		b.WriteString("\n")
		if tr.Color {
			b.WriteString(bold(underline(fr.Path)))
		} else {
			b.WriteString(fr.Path)
		}
		b.WriteString("\n")

		for _, f := range fr.Findings {
			loc := fmt.Sprintf("  %d:%d", f.Line, f.Column)
			sev := f.Severity.String()
			msg := f.Message

			if tr.Color {
				sev = colorSeverity(f.Severity)
			}

			b.WriteString(fmt.Sprintf("  %-8s %-8s %s\n", loc, sev, msg))
		}

		if fr.SyncResult != nil && !fr.SyncResult.InSync() {
			if len(fr.SyncResult.MissingInExample) > 0 {
				b.WriteString("\n  Missing keys (present in .env but not in .env.example):\n")
				for _, k := range fr.SyncResult.MissingInExample {
					b.WriteString(fmt.Sprintf("    - %s\n", k))
				}
			}
			if len(fr.SyncResult.MissingInEnv) > 0 {
				b.WriteString("\n  Missing keys (present in .env.example but not in .env):\n")
				for _, k := range fr.SyncResult.MissingInEnv {
					b.WriteString(fmt.Sprintf("    - %s\n", k))
				}
			}
		}
	}

	errors, warnings, infos := report.Counts()
	total := errors + warnings + infos

	b.WriteString("\n")
	if total == 0 {
		if tr.Color {
			b.WriteString(green("✔ No issues found\n"))
		} else {
			b.WriteString("No issues found\n")
		}
	} else {
		summary := fmt.Sprintf("%d error%s, %d warning%s, %d info",
			errors, pluralS(errors),
			warnings, pluralS(warnings),
			infos,
		)
		if tr.Color {
			if errors > 0 {
				b.WriteString(red("✖ " + summary + "\n"))
			} else if warnings > 0 {
				b.WriteString(yellow("⚠ " + summary + "\n"))
			} else {
				b.WriteString(blue("ℹ " + summary + "\n"))
			}
		} else {
			if errors > 0 {
				b.WriteString("✖ ")
			} else if warnings > 0 {
				b.WriteString("⚠ ")
			} else {
				b.WriteString("ℹ ")
			}
			b.WriteString(summary + "\n")
		}
	}

	return b.String(), nil
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// ANSI color helpers — no external dependencies.
func colorSeverity(s rules.Severity) string {
	switch s {
	case rules.Error:
		return red(s.String())
	case rules.Warning:
		return yellow(s.String())
	case rules.Info:
		return blue(s.String())
	default:
		return s.String()
	}
}

func red(s string) string     { return "\033[31m" + s + "\033[0m" }
func yellow(s string) string  { return "\033[33m" + s + "\033[0m" }
func blue(s string) string    { return "\033[34m" + s + "\033[0m" }
func green(s string) string   { return "\033[32m" + s + "\033[0m" }
func bold(s string) string    { return "\033[1m" + s + "\033[0m" }
func underline(s string) string { return "\033[4m" + s + "\033[0m" }
