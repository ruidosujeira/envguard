package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/romano/envguard/internal/parser"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name             string
		env              string
		example          string
		wantMissExample  []string
		wantMissEnv      []string
		wantInSync       bool
	}{
		{
			name:       "in sync",
			env:        "FOO=bar\nBAZ=qux",
			example:    "FOO=\nBAZ=",
			wantInSync: true,
		},
		{
			name:            "missing in example",
			env:             "FOO=bar\nBAZ=qux\nNEW_KEY=value",
			example:         "FOO=\nBAZ=",
			wantMissExample: []string{"NEW_KEY"},
			wantInSync:      false,
		},
		{
			name:        "missing in env",
			env:         "FOO=bar",
			example:     "FOO=\nBAZ=\nQUX=",
			wantMissEnv: []string{"BAZ", "QUX"},
			wantInSync:  false,
		},
		{
			name:            "both missing",
			env:             "FOO=bar\nNEW=value",
			example:         "FOO=\nOLD=",
			wantMissExample: []string{"NEW"},
			wantMissEnv:     []string{"OLD"},
			wantInSync:      false,
		},
		{
			name:       "empty files",
			env:        "",
			example:    "",
			wantInSync: true,
		},
		{
			name:       "comments and blanks ignored",
			env:        "# comment\nFOO=bar\n\n",
			example:    "# different comment\nFOO=",
			wantInSync: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envFile, err := parser.Parse(strings.NewReader(tt.env))
			if err != nil {
				t.Fatal(err)
			}
			exFile, err := parser.Parse(strings.NewReader(tt.example))
			if err != nil {
				t.Fatal(err)
			}

			result := Compare(envFile, exFile)

			if result.InSync() != tt.wantInSync {
				t.Errorf("InSync() = %v, want %v", result.InSync(), tt.wantInSync)
			}

			if !strSliceEqual(result.MissingInExample, tt.wantMissExample) {
				t.Errorf("MissingInExample = %v, want %v", result.MissingInExample, tt.wantMissExample)
			}

			if !strSliceEqual(result.MissingInEnv, tt.wantMissEnv) {
				t.Errorf("MissingInEnv = %v, want %v", result.MissingInEnv, tt.wantMissEnv)
			}
		})
	}
}

func TestFix(t *testing.T) {
	dir := t.TempDir()
	exPath := filepath.Join(dir, ".env.example")

	// Create initial .env.example.
	if err := os.WriteFile(exPath, []byte("FOO=\nBAR=\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := Fix(exPath, []string{"NEW_KEY", "ANOTHER_KEY"})
	if err != nil {
		t.Fatalf("Fix() error = %v", err)
	}

	content, err := os.ReadFile(exPath)
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "NEW_KEY=") {
		t.Error("expected NEW_KEY= in fixed file")
	}
	if !strings.Contains(s, "ANOTHER_KEY=") {
		t.Error("expected ANOTHER_KEY= in fixed file")
	}
	if !strings.Contains(s, "# Added by envguard sync") {
		t.Error("expected sync comment in fixed file")
	}
}

func TestFixNoKeys(t *testing.T) {
	err := Fix("/nonexistent/path", nil)
	if err != nil {
		t.Errorf("Fix with no keys should not error, got: %v", err)
	}
}

func strSliceEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
