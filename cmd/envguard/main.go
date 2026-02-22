package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/romano/envguard/internal/gitignore"
	"github.com/romano/envguard/internal/parser"
	"github.com/romano/envguard/internal/reporter"
	"github.com/romano/envguard/internal/rules"
	envsync "github.com/romano/envguard/internal/sync"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

var (
	formatFlag string
	fixFlag    bool
	strictFlag bool
	noColorFlag bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "envguard",
	Short:   "Lint, validate, and guard your .env files",
	Long:    "envguard — Your .env files are broken. You just don't know it yet.",
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLint(cmd, args)
	},
}

var lintCmd = &cobra.Command{
	Use:   "lint [files...]",
	Short: "Lint .env files for common issues",
	Long:  "Analyzes .env files for duplicate keys, formatting issues, typos, and more.",
	RunE:  runLint,
}

var secretsCmd = &cobra.Command{
	Use:   "secrets [files...]",
	Short: "Detect possible secrets in .env files",
	RunE:  runSecrets,
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Check sync between .env and .env.example",
	Long:  "Detects variables that exist in .env but not in .env.example, and vice versa.",
	RunE:  runSync,
}

var checkCmd = &cobra.Command{
	Use:   "check [files...]",
	Short: "Run all checks (lint + secrets + sync)",
	RunE:  runCheck,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&formatFlag, "format", "text", "Output format: text, json")
	rootCmd.PersistentFlags().BoolVar(&noColorFlag, "no-color", false, "Disable colored output")
	lintCmd.Flags().BoolVar(&fixFlag, "fix", false, "Auto-fix formatting issues")
	lintCmd.Flags().BoolVar(&strictFlag, "strict", false, "Exit with non-zero code on any finding")
	syncCmd.Flags().BoolVar(&fixFlag, "fix", false, "Auto-add missing keys to .env.example")
	checkCmd.Flags().BoolVar(&strictFlag, "strict", false, "Exit with non-zero code on any finding")

	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(secretsCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(checkCmd)
}

func runLint(cmd *cobra.Command, args []string) error {
	files := resolveFiles(args, ".env")
	report := &reporter.Report{}

	allRules := rules.DefaultRules()

	for _, path := range files {
		envFile, err := parser.ParseFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		findings := rules.RunAll(envFile, allRules)

		// Check gitignore.
		giResult := gitignore.Check(path)
		if !giResult.EnvIgnored {
			findings = append(findings, rules.Finding{
				RuleName: "gitignore",
				Severity: rules.Warning,
				Message:  fmt.Sprintf("%s is not in .gitignore", filepath.Base(path)),
				Line:     0,
				Column:   0,
			})
		}

		report.Files = append(report.Files, reporter.FileReport{
			Path:     path,
			Findings: findings,
		})
	}

	return outputReport(report)
}

func runSecrets(cmd *cobra.Command, args []string) error {
	files := resolveFiles(args, ".env")
	report := &reporter.Report{}

	secretRule := &rules.SecretDetect{}

	for _, path := range files {
		envFile, err := parser.ParseFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		findings := secretRule.Run(envFile)
		report.Files = append(report.Files, reporter.FileReport{
			Path:     path,
			Findings: findings,
		})
	}

	return outputReport(report)
}

func runSync(cmd *cobra.Command, args []string) error {
	envPath := ".env"
	examplePath := ".env.example"

	envFile, err := parser.ParseFile(envPath)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", envPath, err)
	}

	exampleFile, err := parser.ParseFile(examplePath)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", examplePath, err)
	}

	result := envsync.Compare(envFile, exampleFile)

	if fixFlag && len(result.MissingInExample) > 0 {
		if err := envsync.Fix(examplePath, result.MissingInExample); err != nil {
			return fmt.Errorf("fixing %s: %w", examplePath, err)
		}
		fmt.Fprintf(os.Stderr, "Added %d missing key(s) to %s\n", len(result.MissingInExample), examplePath)
	}

	report := &reporter.Report{
		Files: []reporter.FileReport{
			{
				Path:       examplePath,
				SyncResult: result,
			},
		},
	}

	return outputReport(report)
}

func runCheck(cmd *cobra.Command, args []string) error {
	files := resolveFiles(args, ".env")
	report := &reporter.Report{}

	allRules := rules.DefaultRules()
	allRules = append(allRules, &rules.SecretDetect{})

	for _, path := range files {
		envFile, err := parser.ParseFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		findings := rules.RunAll(envFile, allRules)

		// Check gitignore.
		giResult := gitignore.Check(path)
		if !giResult.EnvIgnored {
			findings = append(findings, rules.Finding{
				RuleName: "gitignore",
				Severity: rules.Warning,
				Message:  fmt.Sprintf("%s is not in .gitignore", filepath.Base(path)),
				Line:     0,
				Column:   0,
			})
		}

		report.Files = append(report.Files, reporter.FileReport{
			Path:     path,
			Findings: findings,
		})
	}

	// Also run sync if .env.example exists.
	if _, err := os.Stat(".env.example"); err == nil {
		envFile, err1 := parser.ParseFile(".env")
		exFile, err2 := parser.ParseFile(".env.example")
		if err1 == nil && err2 == nil {
			result := envsync.Compare(envFile, exFile)
			if !result.InSync() {
				report.Files = append(report.Files, reporter.FileReport{
					Path:       ".env.example",
					SyncResult: result,
				})
			}
		}
	}

	return outputReport(report)
}

func outputReport(report *reporter.Report) error {
	var rep reporter.Reporter
	switch formatFlag {
	case "json":
		rep = &reporter.JSONReporter{}
	default:
		rep = &reporter.TextReporter{Color: !noColorFlag && isTerminal()}
	}

	out, err := rep.Report(report)
	if err != nil {
		return err
	}

	fmt.Print(out)

	if report.HasErrors() || (strictFlag && report.HasFindings()) {
		os.Exit(1)
	}

	return nil
}

func resolveFiles(args []string, defaultFile string) []string {
	if len(args) > 0 {
		return args
	}

	// Look for .env files in current directory.
	matches, _ := filepath.Glob(".env*")
	var envFiles []string
	for _, m := range matches {
		base := filepath.Base(m)
		// Skip .env.example and backup files.
		if base == ".env.example" || base == ".env.bak" {
			continue
		}
		envFiles = append(envFiles, m)
	}

	if len(envFiles) == 0 {
		return []string{defaultFile}
	}

	return envFiles
}

// isTerminal returns true if stdout appears to be a terminal.
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
