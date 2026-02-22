package gitignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Result holds the gitignore check result.
type Result struct {
	GitignoreExists bool
	EnvIgnored      bool
	GitignorePath   string
}

// Check verifies if the given env file is listed in .gitignore.
// It looks for a .gitignore in the same directory as the env file,
// then walks up to parent directories.
func Check(envPath string) *Result {
	absEnv, err := filepath.Abs(envPath)
	if err != nil {
		return &Result{}
	}

	envBase := filepath.Base(absEnv)
	dir := filepath.Dir(absEnv)

	// Walk up directories looking for .gitignore.
	for {
		giPath := filepath.Join(dir, ".gitignore")
		if _, err := os.Stat(giPath); err == nil {
			ignored := isFileIgnored(giPath, envBase)
			return &Result{
				GitignoreExists: true,
				EnvIgnored:      ignored,
				GitignorePath:   giPath,
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}

	return &Result{
		GitignoreExists: false,
		EnvIgnored:      false,
	}
}

// isFileIgnored checks if a filename is matched by any pattern in the .gitignore.
func isFileIgnored(gitignorePath, filename string) bool {
	f, err := os.Open(gitignorePath)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle negation patterns.
		if strings.HasPrefix(line, "!") {
			continue // simplified: skip negation
		}

		// Remove trailing slash (directory indicator).
		pattern := strings.TrimSuffix(line, "/")

		// Check for match.
		if matchPattern(pattern, filename) {
			return true
		}
	}

	return false
}

// matchPattern performs simplified gitignore pattern matching.
func matchPattern(pattern, filename string) bool {
	// Exact match.
	if pattern == filename {
		return true
	}

	// Wildcard patterns using filepath.Match.
	matched, err := filepath.Match(pattern, filename)
	if err == nil && matched {
		return true
	}

	// Pattern like ".env*" should match ".env", ".env.local", etc.
	if strings.Contains(pattern, "*") {
		matched, err := filepath.Match(pattern, filename)
		if err == nil && matched {
			return true
		}
	}

	return false
}
