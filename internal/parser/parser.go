package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Entry represents a single line in a .env file.
type Entry struct {
	Key        string
	Value      string
	RawLine    string
	LineNumber int
	IsComment  bool
	IsBlank    bool
	Quoted     rune // 0 if unquoted, '\'' or '"' if quoted
}

// EnvFile represents a parsed .env file.
type EnvFile struct {
	Path    string
	Entries []Entry
}

// Keys returns all non-comment, non-blank entries.
func (f *EnvFile) Keys() []Entry {
	var keys []Entry
	for _, e := range f.Entries {
		if !e.IsComment && !e.IsBlank {
			keys = append(keys, e)
		}
	}
	return keys
}

// KeyNames returns a deduplicated list of key names in order of first appearance.
func (f *EnvFile) KeyNames() []string {
	seen := make(map[string]bool)
	var names []string
	for _, e := range f.Entries {
		if !e.IsComment && !e.IsBlank && !seen[e.Key] {
			seen[e.Key] = true
			names = append(names, e.Key)
		}
	}
	return names
}

// ParseFile parses a .env file from disk.
func ParseFile(path string) (*EnvFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	envFile, err := Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	envFile.Path = path
	return envFile, nil
}

// Parse parses .env content from a reader.
func Parse(r io.Reader) (*EnvFile, error) {
	envFile := &EnvFile{}
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		entry := parseLine(raw, lineNum)
		envFile.Entries = append(envFile.Entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	return envFile, nil
}

func parseLine(raw string, lineNum int) Entry {
	entry := Entry{
		RawLine:    raw,
		LineNumber: lineNum,
	}

	trimmed := strings.TrimSpace(raw)

	if trimmed == "" {
		entry.IsBlank = true
		return entry
	}

	if strings.HasPrefix(trimmed, "#") {
		entry.IsComment = true
		return entry
	}

	// Find the first '=' to split key and value.
	eqIdx := strings.Index(trimmed, "=")
	if eqIdx < 0 {
		// Line has no '=', treat as key with empty value.
		// Keep the raw key (preserving spaces for linting).
		entry.Key = trimmed
		return entry
	}

	// Use raw line (not trimmed) to preserve leading spaces for linting,
	// but for key extraction we work with the portion before '='.
	rawBeforeEq := raw[:strings.Index(raw, "=")]
	entry.Key = rawBeforeEq

	// Extract value after '='.
	rawValue := raw[strings.Index(raw, "=")+1:]

	// Parse the value, handling quotes.
	entry.Value, entry.Quoted = parseValue(rawValue)

	return entry
}

// parseValue extracts the value, respecting single and double quotes.
// Returns the parsed value and the quote character used (0 if unquoted).
func parseValue(raw string) (string, rune) {
	trimmed := strings.TrimSpace(raw)

	if trimmed == "" {
		return "", 0
	}

	// Check for quoted values.
	if len(trimmed) >= 2 {
		first := rune(trimmed[0])
		if (first == '"' || first == '\'') && rune(trimmed[len(trimmed)-1]) == first {
			// Remove quotes and return inner value.
			inner := trimmed[1 : len(trimmed)-1]
			if first == '"' {
				// Handle escape sequences in double-quoted values.
				inner = expandEscapes(inner)
			}
			return inner, first
		}
	}

	// Unquoted: strip inline comments (# preceded by whitespace).
	if idx := findInlineComment(trimmed); idx >= 0 {
		trimmed = strings.TrimRight(trimmed[:idx], " \t")
	}

	return trimmed, 0
}

// expandEscapes handles common escape sequences in double-quoted values.
func expandEscapes(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	escaped := false
	for _, r := range s {
		if escaped {
			switch r {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case '\\':
				b.WriteByte('\\')
			case '"':
				b.WriteByte('"')
			default:
				b.WriteByte('\\')
				b.WriteRune(r)
			}
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		b.WriteRune(r)
	}
	if escaped {
		b.WriteByte('\\')
	}
	return b.String()
}

// findInlineComment returns the index of an inline comment (# preceded by whitespace)
// or -1 if none found.
func findInlineComment(s string) int {
	for i := 1; i < len(s); i++ {
		if s[i] == '#' && (s[i-1] == ' ' || s[i-1] == '\t') {
			return i
		}
	}
	return -1
}
