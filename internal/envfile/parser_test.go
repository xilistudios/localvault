package envfile

import (
	"strings"
	"testing"
)

func TestParseSimple(t *testing.T) {
	input := "KEY=value"
	pairs, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0].Key != "KEY" || pairs[0].Value != "value" {
		t.Errorf("expected KEY=value, got %s=%s", pairs[0].Key, pairs[0].Value)
	}
}

func TestParseMultipleLines(t *testing.T) {
	input := "A=1\nB=2\nC=3"
	pairs, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 3 {
		t.Fatalf("expected 3 pairs, got %d", len(pairs))
	}
	expected := map[string]string{"A": "1", "B": "2", "C": "3"}
	for _, p := range pairs {
		if expected[p.Key] != p.Value {
			t.Errorf("expected %s=%s, got %s", p.Key, expected[p.Key], p.Value)
		}
	}
}

func TestParseCommentsAndEmptyLines(t *testing.T) {
	input := "# comment\n\nKEY=value\n\n# another comment\nOTHER=thing"
	pairs, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 2 {
		t.Fatalf("expected 2 pairs, got %d", len(pairs))
	}
	if pairs[0].Key != "KEY" || pairs[1].Key != "OTHER" {
		t.Errorf("unexpected keys: %s, %s", pairs[0].Key, pairs[1].Key)
	}
}

func TestParseExportPrefix(t *testing.T) {
	input := "export DATABASE_URL=postgres://localhost/mydb"
	pairs, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(pairs))
	}
	if pairs[0].Key != "DATABASE_URL" {
		t.Errorf("expected key DATABASE_URL, got %s", pairs[0].Key)
	}
	if pairs[0].Value != "postgres://localhost/mydb" {
		t.Errorf("expected value postgres://localhost/mydb, got %s", pairs[0].Value)
	}
}

func TestParseDoubleQuotedValue(t *testing.T) {
	input := `KEY="hello world"`
	pairs, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pairs[0].Value != "hello world" {
		t.Errorf("expected 'hello world', got %q", pairs[0].Value)
	}
}

func TestParseSingleQuotedValue(t *testing.T) {
	input := "KEY='hello world'"
	pairs, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pairs[0].Value != "hello world" {
		t.Errorf("expected 'hello world', got %q", pairs[0].Value)
	}
}

func TestParseValueWithEquals(t *testing.T) {
	input := "KEY=postgres://user:pass@host/db?opt=1"
	pairs, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "postgres://user:pass@host/db?opt=1"
	if pairs[0].Value != expected {
		t.Errorf("expected %q, got %q", expected, pairs[0].Value)
	}
}

func TestParseInvalidLine(t *testing.T) {
	input := "INVALID"
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for invalid line, got nil")
	}
}

func TestParseEmptyInput(t *testing.T) {
	pairs, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pairs) != 0 {
		t.Errorf("expected 0 pairs, got %d", len(pairs))
	}
}

func TestToMap(t *testing.T) {
	pairs := []Pair{
		{Key: "A", Value: "1"},
		{Key: "B", Value: "2"},
	}
	m := ToMap(pairs)
	if len(m) != 2 {
		t.Fatalf("expected map of size 2, got %d", len(m))
	}
	if m["A"] != "1" || m["B"] != "2" {
		t.Errorf("unexpected map values: %v", m)
	}
}
