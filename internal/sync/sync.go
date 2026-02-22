package sync

import (
	"fmt"
	"os"
	"strings"

	"github.com/romano/envguard/internal/parser"
)

// Result holds the sync comparison between two env files.
type Result struct {
	// MissingInExample are keys present in .env but not in .env.example.
	MissingInExample []string
	// MissingInEnv are keys present in .env.example but not in .env.
	MissingInEnv []string
}

// InSync returns true if both files have the same keys.
func (r *Result) InSync() bool {
	return len(r.MissingInExample) == 0 && len(r.MissingInEnv) == 0
}

// Compare compares the keys in two parsed env files.
func Compare(env, example *parser.EnvFile) *Result {
	envKeys := keySet(env)
	exampleKeys := keySet(example)

	result := &Result{}

	for _, name := range env.KeyNames() {
		if !exampleKeys[name] {
			result.MissingInExample = append(result.MissingInExample, name)
		}
	}

	for _, name := range example.KeyNames() {
		if !envKeys[name] {
			result.MissingInEnv = append(result.MissingInEnv, name)
		}
	}

	return result
}

// Fix adds missing keys to the .env.example file.
// Keys from .env that are missing in .env.example are appended with empty values.
func Fix(examplePath string, missingKeys []string) error {
	if len(missingKeys) == 0 {
		return nil
	}

	f, err := os.OpenFile(examplePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening %s: %w", examplePath, err)
	}
	defer f.Close()

	var b strings.Builder
	b.WriteString("\n# Added by envguard sync\n")
	for _, key := range missingKeys {
		b.WriteString(key)
		b.WriteString("=\n")
	}

	if _, err := f.WriteString(b.String()); err != nil {
		return fmt.Errorf("writing to %s: %w", examplePath, err)
	}

	return nil
}

func keySet(file *parser.EnvFile) map[string]bool {
	set := make(map[string]bool)
	for _, name := range file.KeyNames() {
		set[name] = true
	}
	return set
}
