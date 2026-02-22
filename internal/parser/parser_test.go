package parser

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []Entry
	}{
		{
			name:  "simple key=value",
			input: "FOO=bar",
			want: []Entry{
				{Key: "FOO", Value: "bar", RawLine: "FOO=bar", LineNumber: 1},
			},
		},
		{
			name:  "double quoted value",
			input: `FOO="bar baz"`,
			want: []Entry{
				{Key: "FOO", Value: "bar baz", RawLine: `FOO="bar baz"`, LineNumber: 1, Quoted: '"'},
			},
		},
		{
			name:  "single quoted value",
			input: `FOO='bar baz'`,
			want: []Entry{
				{Key: "FOO", Value: "bar baz", RawLine: `FOO='bar baz'`, LineNumber: 1, Quoted: '\''},
			},
		},
		{
			name:  "empty value",
			input: "FOO=",
			want: []Entry{
				{Key: "FOO", Value: "", RawLine: "FOO=", LineNumber: 1},
			},
		},
		{
			name:  "comment line",
			input: "# this is a comment",
			want: []Entry{
				{RawLine: "# this is a comment", LineNumber: 1, IsComment: true},
			},
		},
		{
			name:  "blank line",
			input: "FOO=bar\n\nBAZ=qux",
			want: []Entry{
				{Key: "FOO", Value: "bar", RawLine: "FOO=bar", LineNumber: 1},
				{RawLine: "", LineNumber: 2, IsBlank: true},
				{Key: "BAZ", Value: "qux", RawLine: "BAZ=qux", LineNumber: 3},
			},
		},
		{
			name:  "value with inline comment",
			input: "FOO=bar # comment",
			want: []Entry{
				{Key: "FOO", Value: "bar", RawLine: "FOO=bar # comment", LineNumber: 1},
			},
		},
		{
			name:  "quoted value preserves hash",
			input: `FOO="bar # not a comment"`,
			want: []Entry{
				{Key: "FOO", Value: "bar # not a comment", RawLine: `FOO="bar # not a comment"`, LineNumber: 1, Quoted: '"'},
			},
		},
		{
			name:  "escape sequences in double quotes",
			input: `FOO="hello\nworld"`,
			want: []Entry{
				{Key: "FOO", Value: "hello\nworld", RawLine: `FOO="hello\nworld"`, LineNumber: 1, Quoted: '"'},
			},
		},
		{
			name:  "no escape in single quotes",
			input: `FOO='hello\nworld'`,
			want: []Entry{
				{Key: "FOO", Value: `hello\nworld`, RawLine: `FOO='hello\nworld'`, LineNumber: 1, Quoted: '\''},
			},
		},
		{
			name:  "key without value or equals",
			input: "FOO",
			want: []Entry{
				{Key: "FOO", Value: "", RawLine: "FOO", LineNumber: 1},
			},
		},
		{
			name:  "spaces around key preserved in Key field",
			input: " FOO =bar",
			want: []Entry{
				{Key: " FOO ", Value: "bar", RawLine: " FOO =bar", LineNumber: 1},
			},
		},
		{
			name:  "value with equals sign",
			input: "DATABASE_URL=postgres://user:pass@host/db?sslmode=require",
			want: []Entry{
				{Key: "DATABASE_URL", Value: "postgres://user:pass@host/db?sslmode=require", RawLine: "DATABASE_URL=postgres://user:pass@host/db?sslmode=require", LineNumber: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(got.Entries) != len(tt.want) {
				t.Fatalf("got %d entries, want %d", len(got.Entries), len(tt.want))
			}

			for i, want := range tt.want {
				g := got.Entries[i]
				if g.Key != want.Key {
					t.Errorf("entry[%d].Key = %q, want %q", i, g.Key, want.Key)
				}
				if g.Value != want.Value {
					t.Errorf("entry[%d].Value = %q, want %q", i, g.Value, want.Value)
				}
				if g.RawLine != want.RawLine {
					t.Errorf("entry[%d].RawLine = %q, want %q", i, g.RawLine, want.RawLine)
				}
				if g.LineNumber != want.LineNumber {
					t.Errorf("entry[%d].LineNumber = %d, want %d", i, g.LineNumber, want.LineNumber)
				}
				if g.IsComment != want.IsComment {
					t.Errorf("entry[%d].IsComment = %v, want %v", i, g.IsComment, want.IsComment)
				}
				if g.IsBlank != want.IsBlank {
					t.Errorf("entry[%d].IsBlank = %v, want %v", i, g.IsBlank, want.IsBlank)
				}
				if g.Quoted != want.Quoted {
					t.Errorf("entry[%d].Quoted = %q, want %q", i, g.Quoted, want.Quoted)
				}
			}
		})
	}
}

func TestParseMultiLine(t *testing.T) {
	input := `# Database config
DATABASE_URL=postgres://localhost/mydb
REDIS_URL=

# API keys
API_KEY="sk-1234567890"
SECRET=
`

	got, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(got.Entries) != 7 {
		t.Fatalf("got %d entries, want 7", len(got.Entries))
	}

	keys := got.Keys()
	if len(keys) != 4 {
		t.Fatalf("got %d keys, want 4", len(keys))
	}

	names := got.KeyNames()
	expected := []string{"DATABASE_URL", "REDIS_URL", "API_KEY", "SECRET"}
	if len(names) != len(expected) {
		t.Fatalf("got %d names, want %d", len(names), len(expected))
	}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("name[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestKeyNames_Deduplicates(t *testing.T) {
	input := "FOO=1\nBAR=2\nFOO=3"
	got, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	names := got.KeyNames()
	if len(names) != 2 {
		t.Fatalf("got %d names, want 2 (FOO, BAR)", len(names))
	}
	if names[0] != "FOO" || names[1] != "BAR" {
		t.Errorf("names = %v, want [FOO BAR]", names)
	}
}
