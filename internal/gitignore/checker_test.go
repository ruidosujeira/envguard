package gitignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name            string
		gitignore       string // content of .gitignore, empty string means no .gitignore
		envFile         string // name of the env file
		wantExists      bool
		wantIgnored     bool
		createGitignore bool
	}{
		{
			name:            ".env in gitignore",
			gitignore:       ".env\nnode_modules/\n",
			envFile:         ".env",
			wantExists:      true,
			wantIgnored:     true,
			createGitignore: true,
		},
		{
			name:            ".env not in gitignore",
			gitignore:       "node_modules/\n*.log\n",
			envFile:         ".env",
			wantExists:      true,
			wantIgnored:     false,
			createGitignore: true,
		},
		{
			name:            "wildcard .env*",
			gitignore:       ".env*\n",
			envFile:         ".env.local",
			wantExists:      true,
			wantIgnored:     true,
			createGitignore: true,
		},
		{
			name:            "no gitignore",
			envFile:         ".env",
			wantExists:      false,
			wantIgnored:     false,
			createGitignore: false,
		},
		{
			name:            "comment lines ignored",
			gitignore:       "# .env\nnode_modules/\n",
			envFile:         ".env",
			wantExists:      true,
			wantIgnored:     false,
			createGitignore: true,
		},
		{
			name:            "blank lines ignored",
			gitignore:       "\n\n.env\n\n",
			envFile:         ".env",
			wantExists:      true,
			wantIgnored:     true,
			createGitignore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.createGitignore {
				giPath := filepath.Join(dir, ".gitignore")
				if err := os.WriteFile(giPath, []byte(tt.gitignore), 0644); err != nil {
					t.Fatal(err)
				}
			}

			envPath := filepath.Join(dir, tt.envFile)
			if err := os.WriteFile(envPath, []byte("FOO=bar\n"), 0644); err != nil {
				t.Fatal(err)
			}

			result := Check(envPath)

			if result.GitignoreExists != tt.wantExists {
				t.Errorf("GitignoreExists = %v, want %v", result.GitignoreExists, tt.wantExists)
			}
			if result.EnvIgnored != tt.wantIgnored {
				t.Errorf("EnvIgnored = %v, want %v", result.EnvIgnored, tt.wantIgnored)
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern  string
		filename string
		want     bool
	}{
		{".env", ".env", true},
		{".env*", ".env", true},
		{".env*", ".env.local", true},
		{"*.log", "app.log", true},
		{"node_modules", ".env", false},
		{".env", ".env.local", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.filename, func(t *testing.T) {
			got := matchPattern(tt.pattern, tt.filename)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.pattern, tt.filename, got, tt.want)
			}
		})
	}
}
