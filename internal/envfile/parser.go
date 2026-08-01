package envfile

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Pair is a single key-value from a .env file
type Pair struct {
	Key   string
	Value string
}

// Parse reads a .env format input and returns key-value pairs.
// Preserves insertion order via slice of pairs.
func Parse(r io.Reader) ([]Pair, error) {
	var pairs []Pair
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Strip "export " prefix
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimPrefix(line, "export ")
			line = strings.TrimSpace(line)
		}

		// Split on first =
		idx := strings.Index(line, "=")
		if idx <= 0 {
			return nil, fmt.Errorf("line %d: invalid format (expected KEY=VALUE): %q", lineNum, line)
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		// Strip surrounding quotes
		value = stripQuotes(value)

		pairs = append(pairs, Pair{Key: key, Value: value})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	return pairs, nil
}

// ToMap converts pairs to a map
func ToMap(pairs []Pair) map[string]string {
	m := make(map[string]string, len(pairs))
	for _, p := range pairs {
		m[p.Key] = p.Value
	}
	return m
}

func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
