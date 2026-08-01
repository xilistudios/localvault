package envfile

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestExportDotenv(t *testing.T) {
	secrets := map[string]string{"B": "2", "A": "1"}
	out, err := Export(secrets, FormatDotenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be sorted
	expected := "A=1\nB=2\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestExportJSON(t *testing.T) {
	secrets := map[string]string{"KEY": "val", "OTHER": "thing"}
	out, err := Export(secrets, FormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]string
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["KEY"] != "val" || parsed["OTHER"] != "thing" {
		t.Errorf("unexpected JSON content: %v", parsed)
	}
}

func TestExportDocker(t *testing.T) {
	secrets := map[string]string{"A": "1", "B": "2"}
	out, err := Export(secrets, FormatDocker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "-e A=1\n-e B=2\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestExportUnknownFormat(t *testing.T) {
	secrets := map[string]string{"A": "1"}
	_, err := Export(secrets, Format("yaml"))
	if err == nil {
		t.Fatal("expected error for unknown format, got nil")
	}
	if !strings.Contains(err.Error(), "unknown export format") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExportEmptyMap(t *testing.T) {
	secrets := map[string]string{}

	out, err := Export(secrets, FormatDotenv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("expected empty string, got %q", out)
	}

	out, err = Export(secrets, FormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "{}" {
		t.Errorf("expected '{}', got %q", out)
	}

	out, err = Export(secrets, FormatDocker)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("expected empty string, got %q", out)
	}
}
