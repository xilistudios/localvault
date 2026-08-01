package model

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// NewVaultFile
// ---------------------------------------------------------------------------

func TestNewVaultFile(t *testing.T) {
	vf := NewVaultFile()
	if vf.Version != 1 {
		t.Errorf("expected version 1, got %d", vf.Version)
	}
	if vf.Projects == nil {
		t.Fatal("Projects map should not be nil")
	}
	if len(vf.Projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(vf.Projects))
	}
}

// ---------------------------------------------------------------------------
// ValidateProjectName
// ---------------------------------------------------------------------------

func TestValidateProjectName(t *testing.T) {
	valid := []string{"myapp", "my-app", "app_1", "a", "0", "a-0_b"}
	for _, name := range valid {
		if err := ValidateProjectName(name); err != nil {
			t.Errorf("ValidateProjectName(%q) should pass, got: %v", name, err)
		}
	}

	invalid := []string{"MyApp", "my app", "", strings.Repeat("a", 65), "my.app", "UPPER", "hello/world"}
	for _, name := range invalid {
		if err := ValidateProjectName(name); err == nil {
			t.Errorf("ValidateProjectName(%q) should fail", name)
		}
	}
}

// ---------------------------------------------------------------------------
// ValidateConfigName
// ---------------------------------------------------------------------------

func TestValidateConfigName(t *testing.T) {
	valid := []string{"staging", "prod-01", "dev_env", "a"}
	for _, name := range valid {
		if err := ValidateConfigName(name); err != nil {
			t.Errorf("ValidateConfigName(%q) should pass, got: %v", name, err)
		}
	}

	invalid := []string{"Staging", "prod env", "", strings.Repeat("x", 65), "prod.env"}
	for _, name := range invalid {
		if err := ValidateConfigName(name); err == nil {
			t.Errorf("ValidateConfigName(%q) should fail", name)
		}
	}
}

// ---------------------------------------------------------------------------
// ValidateSecretKey
// ---------------------------------------------------------------------------

func TestValidateSecretKey(t *testing.T) {
	valid := []string{"DATABASE_URL", "A", "KEY_1", "LONG_KEY"}
	for _, key := range valid {
		if err := ValidateSecretKey(key); err != nil {
			t.Errorf("ValidateSecretKey(%q) should pass, got: %v", key, err)
		}
	}

	invalid := []string{"database_url", "my-key", "", strings.Repeat("A", 257), "lower", "Key"}
	for _, key := range invalid {
		if err := ValidateSecretKey(key); err == nil {
			t.Errorf("ValidateSecretKey(%q) should fail", key)
		}
	}
}

// ---------------------------------------------------------------------------
// ValidateSecretValue
// ---------------------------------------------------------------------------

func TestValidateSecretValue(t *testing.T) {
	if err := ValidateSecretValue(""); err != nil {
		t.Errorf("empty string should be valid, got: %v", err)
	}
	if err := ValidateSecretValue("normal value"); err != nil {
		t.Errorf("normal value should be valid, got: %v", err)
	}

	big := strings.Repeat("x", 64*1024+1)
	if err := ValidateSecretValue(big); err == nil {
		t.Error("value exceeding 64KB should be invalid")
	}
}

// ---------------------------------------------------------------------------
// AddProject
// ---------------------------------------------------------------------------

func TestAddProject(t *testing.T) {
	vf := NewVaultFile()

	if err := vf.AddProject("myapp"); err != nil {
		t.Fatalf("AddProject failed: %v", err)
	}
	if _, ok := vf.Projects["myapp"]; !ok {
		t.Fatal("project should exist after AddProject")
	}

	// Duplicate
	if err := vf.AddProject("myapp"); err == nil {
		t.Error("adding duplicate project should fail")
	}

	// Invalid name
	if err := vf.AddProject("INVALID"); err == nil {
		t.Error("adding project with invalid name should fail")
	}
}

// ---------------------------------------------------------------------------
// DeleteProject
// ---------------------------------------------------------------------------

func TestDeleteProject(t *testing.T) {
	vf := NewVaultFile()
	vf.AddProject("app")
	vf.ActiveProject = "app"
	vf.ActiveConfig = "cfg"

	if err := vf.DeleteProject("app"); err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}
	if vf.ActiveProject != "" {
		t.Error("ActiveProject should be cleared after deleting active project")
	}
	if vf.ActiveConfig != "" {
		t.Error("ActiveConfig should be cleared after deleting active project")
	}

	// Not found
	if err := vf.DeleteProject("nonexistent"); err == nil {
		t.Error("deleting nonexistent project should fail")
	}
}

// ---------------------------------------------------------------------------
// AddConfig
// ---------------------------------------------------------------------------

func TestAddConfig(t *testing.T) {
	vf := NewVaultFile()
	vf.AddProject("proj")

	if err := vf.AddConfig("proj", "staging"); err != nil {
		t.Fatalf("AddConfig failed: %v", err)
	}

	// Duplicate
	if err := vf.AddConfig("proj", "staging"); err == nil {
		t.Error("adding duplicate config should fail")
	}

	// Invalid name
	if err := vf.AddConfig("proj", "INVALID"); err == nil {
		t.Error("adding config with invalid name should fail")
	}

	// Project not found
	if err := vf.AddConfig("missing", "cfg"); err == nil {
		t.Error("adding config to missing project should fail")
	}
}

// ---------------------------------------------------------------------------
// DeleteConfig
// ---------------------------------------------------------------------------

func TestDeleteConfig(t *testing.T) {
	vf := NewVaultFile()
	vf.AddProject("proj")
	vf.AddConfig("proj", "prod")
	vf.ActiveProject = "proj"
	vf.ActiveConfig = "prod"

	if err := vf.DeleteConfig("proj", "prod"); err != nil {
		t.Fatalf("DeleteConfig failed: %v", err)
	}
	if vf.ActiveConfig != "" {
		t.Error("ActiveConfig should be cleared after deleting active config")
	}

	// Not found
	if err := vf.DeleteConfig("proj", "missing"); err == nil {
		t.Error("deleting nonexistent config should fail")
	}

	// Project not found
	if err := vf.DeleteConfig("missing", "cfg"); err == nil {
		t.Error("deleting config from missing project should fail")
	}
}

// ---------------------------------------------------------------------------
// GetConfig
// ---------------------------------------------------------------------------

func TestGetConfig(t *testing.T) {
	vf := NewVaultFile()
	vf.AddProject("proj")
	vf.AddConfig("proj", "dev")

	cfg, err := vf.GetConfig("proj", "dev")
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("config should not be nil")
	}

	// Project not found
	if _, err := vf.GetConfig("missing", "dev"); err == nil {
		t.Error("GetConfig should fail for missing project")
	}

	// Config not found
	if _, err := vf.GetConfig("proj", "missing"); err == nil {
		t.Error("GetConfig should fail for missing config")
	}
}

// ---------------------------------------------------------------------------
// AddSecretToIndex
// ---------------------------------------------------------------------------

func TestAddSecretToIndex(t *testing.T) {
	vf := NewVaultFile()
	vf.AddProject("proj")
	vf.AddConfig("proj", "dev")

	if err := vf.AddSecretToIndex("proj", "dev", "DATABASE_URL"); err != nil {
		t.Fatalf("AddSecretToIndex failed: %v", err)
	}
	cfg, _ := vf.GetConfig("proj", "dev")
	if len(cfg.Secrets) != 1 || cfg.Secrets[0] != "DATABASE_URL" {
		t.Fatalf("expected [DATABASE_URL], got %v", cfg.Secrets)
	}

	// Idempotent
	if err := vf.AddSecretToIndex("proj", "dev", "DATABASE_URL"); err != nil {
		t.Fatalf("idempotent AddSecretToIndex failed: %v", err)
	}
	if len(cfg.Secrets) != 1 {
		t.Fatalf("expected 1 secret after idempotent add, got %d", len(cfg.Secrets))
	}
}

// ---------------------------------------------------------------------------
// RemoveSecretFromIndex
// ---------------------------------------------------------------------------

func TestRemoveSecretFromIndex(t *testing.T) {
	vf := NewVaultFile()
	vf.AddProject("proj")
	vf.AddConfig("proj", "dev")
	vf.AddSecretToIndex("proj", "dev", "API_KEY")
	vf.AddSecretToIndex("proj", "dev", "DB_PASS")

	if err := vf.RemoveSecretFromIndex("proj", "dev", "API_KEY"); err != nil {
		t.Fatalf("RemoveSecretFromIndex failed: %v", err)
	}
	cfg, _ := vf.GetConfig("proj", "dev")
	if len(cfg.Secrets) != 1 || cfg.Secrets[0] != "DB_PASS" {
		t.Fatalf("expected [DB_PASS], got %v", cfg.Secrets)
	}

	// No-op on missing key
	if err := vf.RemoveSecretFromIndex("proj", "dev", "NONEXISTENT"); err != nil {
		t.Fatalf("removing nonexistent key should be no-op, got: %v", err)
	}
}
