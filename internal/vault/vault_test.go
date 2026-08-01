package vault

import (
	"testing"
)

// helper: init vault in a temp dir and return a *Vault with MockKeyring
func setupTestVault(t *testing.T) *Vault {
	t.Helper()
	t.Setenv("LOCALVAULT_DIR", t.TempDir())
	if err := InitVault(); err != nil {
		t.Fatalf("InitVault() error: %v", err)
	}
	mk := NewMockKeyring()
	v, err := NewVault(mk)
	if err != nil {
		t.Fatalf("NewVault() error: %v", err)
	}
	return v
}

// --- NewVault ---

func TestNewVault_NotInitialized(t *testing.T) {
	t.Setenv("LOCALVAULT_DIR", t.TempDir())
	// Do NOT call InitVault
	_, err := NewVault(NewMockKeyring())
	if err == nil {
		t.Fatal("NewVault() should error when vault not initialized")
	}
}

func TestNewVault_Success(t *testing.T) {
	v := setupTestVault(t)
	if v.Meta == nil {
		t.Fatal("Vault.Meta should not be nil")
	}
	if v.Store == nil {
		t.Fatal("Vault.Store should not be nil")
	}
}

// --- Projects ---

func TestCreateAndListProjects(t *testing.T) {
	v := setupTestVault(t)

	if err := v.CreateProject("beta"); err != nil {
		t.Fatalf("CreateProject(beta) error: %v", err)
	}
	if err := v.CreateProject("alpha"); err != nil {
		t.Fatalf("CreateProject(alpha) error: %v", err)
	}

	list := v.ListProjects()
	if len(list) != 2 {
		t.Fatalf("ListProjects() returned %d, want 2", len(list))
	}
	if list[0] != "alpha" || list[1] != "beta" {
		t.Errorf("ListProjects() = %v, want [alpha beta]", list)
	}
}

func TestDeleteProject(t *testing.T) {
	v := setupTestVault(t)

	// Create project with config and secret
	if err := v.CreateProject("myapp"); err != nil {
		t.Fatal(err)
	}
	if err := v.CreateConfig("myapp", "dev"); err != nil {
		t.Fatal(err)
	}
	if err := v.SetSecret("myapp", "dev", "API_KEY", "secret123"); err != nil {
		t.Fatal(err)
	}

	// Verify secret exists in store
	key := SecretKey("myapp", "dev", "API_KEY")
	if _, err := v.Store.Get(key); err != nil {
		t.Fatalf("secret should exist before delete: %v", err)
	}

	// Delete project
	if err := v.DeleteProject("myapp"); err != nil {
		t.Fatalf("DeleteProject() error: %v", err)
	}

	// Verify secret removed from store
	if _, err := v.Store.Get(key); err == nil {
		t.Fatal("secret should be removed from store after project delete")
	}

	// Verify project gone from metadata
	if projects := v.ListProjects(); len(projects) != 0 {
		t.Errorf("expected no projects, got %v", projects)
	}
}

// --- Configs ---

func TestCreateAndListConfigs(t *testing.T) {
	v := setupTestVault(t)

	if err := v.CreateProject("myapp"); err != nil {
		t.Fatal(err)
	}
	if err := v.CreateConfig("myapp", "prod"); err != nil {
		t.Fatal(err)
	}
	if err := v.CreateConfig("myapp", "dev"); err != nil {
		t.Fatal(err)
	}

	list, err := v.ListConfigs("myapp")
	if err != nil {
		t.Fatalf("ListConfigs() error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("ListConfigs() returned %d, want 2", len(list))
	}
	if list[0] != "dev" || list[1] != "prod" {
		t.Errorf("ListConfigs() = %v, want [dev prod]", list)
	}
}

func TestDeleteConfig(t *testing.T) {
	v := setupTestVault(t)

	if err := v.CreateProject("myapp"); err != nil {
		t.Fatal(err)
	}
	if err := v.CreateConfig("myapp", "dev"); err != nil {
		t.Fatal(err)
	}
	if err := v.SetSecret("myapp", "dev", "API_KEY", "key1"); err != nil {
		t.Fatal(err)
	}
	if err := v.SetSecret("myapp", "dev", "DB_PASS", "pass1"); err != nil {
		t.Fatal(err)
	}

	// Delete config
	if err := v.DeleteConfig("myapp", "dev"); err != nil {
		t.Fatalf("DeleteConfig() error: %v", err)
	}

	// Verify secrets removed from store
	for _, name := range []string{"API_KEY", "DB_PASS"} {
		key := SecretKey("myapp", "dev", name)
		if _, err := v.Store.Get(key); err == nil {
			t.Errorf("secret %s should be removed from store after config delete", name)
		}
	}

	// Verify config gone
	if _, err := v.ListConfigs("myapp"); err != nil {
		t.Fatalf("ListConfigs() error: %v", err)
	}
	cfgs, _ := v.ListConfigs("myapp")
	if len(cfgs) != 0 {
		t.Errorf("expected no configs, got %v", cfgs)
	}
}

// --- CopyConfig ---

func TestCopyConfig(t *testing.T) {
	v := setupTestVault(t)

	if err := v.CreateProject("myapp"); err != nil {
		t.Fatal(err)
	}
	if err := v.CreateConfig("myapp", "dev"); err != nil {
		t.Fatal(err)
	}
	if err := v.SetSecret("myapp", "dev", "API_KEY", "key123"); err != nil {
		t.Fatal(err)
	}
	if err := v.SetSecret("myapp", "dev", "DB_URL", "postgres://localhost"); err != nil {
		t.Fatal(err)
	}

	// Copy dev -> staging
	if err := v.CopyConfig("myapp", "dev", "staging"); err != nil {
		t.Fatalf("CopyConfig() error: %v", err)
	}

	// Verify staging has same secrets
	secrets, err := v.ListSecrets("myapp", "staging")
	if err != nil {
		t.Fatalf("ListSecrets(staging) error: %v", err)
	}
	if len(secrets) != 2 {
		t.Fatalf("expected 2 secrets in staging, got %d", len(secrets))
	}

	// Verify values match
	for _, name := range []string{"API_KEY", "DB_URL"} {
		orig, _ := v.GetSecret("myapp", "dev", name)
		copied, _ := v.GetSecret("myapp", "staging", name)
		if orig != copied {
			t.Errorf("secret %s: dev=%q, staging=%q", name, orig, copied)
		}
	}
}

// --- SetActiveScope ---

func TestSetActiveScope(t *testing.T) {
	v := setupTestVault(t)

	if err := v.CreateProject("myapp"); err != nil {
		t.Fatal(err)
	}
	if err := v.CreateConfig("myapp", "dev"); err != nil {
		t.Fatal(err)
	}

	if err := v.SetActiveScope("myapp", "dev"); err != nil {
		t.Fatalf("SetActiveScope() error: %v", err)
	}

	if v.Meta.ActiveProject != "myapp" {
		t.Errorf("ActiveProject = %q, want %q", v.Meta.ActiveProject, "myapp")
	}
	if v.Meta.ActiveConfig != "dev" {
		t.Errorf("ActiveConfig = %q, want %q", v.Meta.ActiveConfig, "dev")
	}
}

// --- Secrets ---

func TestSetAndGetSecret(t *testing.T) {
	v := setupTestVault(t)

	if err := v.CreateProject("myapp"); err != nil {
		t.Fatal(err)
	}
	if err := v.CreateConfig("myapp", "dev"); err != nil {
		t.Fatal(err)
	}

	want := "postgres://user:pass@localhost:5432/mydb"
	if err := v.SetSecret("myapp", "dev", "DATABASE_URL", want); err != nil {
		t.Fatalf("SetSecret() error: %v", err)
	}

	got, err := v.GetSecret("myapp", "dev", "DATABASE_URL")
	if err != nil {
		t.Fatalf("GetSecret() error: %v", err)
	}
	if got != want {
		t.Errorf("GetSecret() = %q, want %q", got, want)
	}
}

func TestUnsetSecret(t *testing.T) {
	v := setupTestVault(t)

	if err := v.CreateProject("myapp"); err != nil {
		t.Fatal(err)
	}
	if err := v.CreateConfig("myapp", "dev"); err != nil {
		t.Fatal(err)
	}
	if err := v.SetSecret("myapp", "dev", "API_KEY", "secret"); err != nil {
		t.Fatal(err)
	}

	if err := v.UnsetSecret("myapp", "dev", "API_KEY"); err != nil {
		t.Fatalf("UnsetSecret() error: %v", err)
	}

	_, err := v.GetSecret("myapp", "dev", "API_KEY")
	if err == nil {
		t.Fatal("GetSecret() after UnsetSecret() should return error")
	}
}

func TestListSecrets(t *testing.T) {
	v := setupTestVault(t)

	if err := v.CreateProject("myapp"); err != nil {
		t.Fatal(err)
	}
	if err := v.CreateConfig("myapp", "dev"); err != nil {
		t.Fatal(err)
	}

	for _, k := range []string{"Z_KEY", "A_KEY", "M_KEY"} {
		if err := v.SetSecret("myapp", "dev", k, "val"); err != nil {
			t.Fatalf("SetSecret(%s) error: %v", k, err)
		}
	}

	secrets, err := v.ListSecrets("myapp", "dev")
	if err != nil {
		t.Fatalf("ListSecrets() error: %v", err)
	}

	want := []string{"A_KEY", "M_KEY", "Z_KEY"}
	if len(secrets) != len(want) {
		t.Fatalf("ListSecrets() returned %d, want %d", len(secrets), len(want))
	}
	for i, s := range secrets {
		if s != want[i] {
			t.Errorf("ListSecrets()[%d] = %q, want %q", i, s, want[i])
		}
	}
}

func TestGetAllSecrets(t *testing.T) {
	v := setupTestVault(t)

	if err := v.CreateProject("myapp"); err != nil {
		t.Fatal(err)
	}
	if err := v.CreateConfig("myapp", "dev"); err != nil {
		t.Fatal(err)
	}

	secrets := map[string]string{
		"DATABASE_URL": "postgres://localhost",
		"API_KEY":      "key123",
		"SECRET_TOKEN": "tok456",
	}
	for k, val := range secrets {
		if err := v.SetSecret("myapp", "dev", k, val); err != nil {
			t.Fatalf("SetSecret(%s) error: %v", k, err)
		}
	}

	got, err := v.GetAllSecrets("myapp", "dev")
	if err != nil {
		t.Fatalf("GetAllSecrets() error: %v", err)
	}
	if len(got) != len(secrets) {
		t.Fatalf("GetAllSecrets() returned %d, want %d", len(got), len(secrets))
	}
	for k, want := range secrets {
		if got[k] != want {
			t.Errorf("GetAllSecrets()[%s] = %q, want %q", k, got[k], want)
		}
	}
}

// --- MaskValue ---

func TestMaskValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "****"},
		{"two chars", "ab", "****"},
		{"four chars", "abcd", "****"},
		{"five chars", "abcde", "ab*de"},
		{"postgres url", "postgres://localhost", "po****************st"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskValue(tt.input)
			if got != tt.want {
				t.Errorf("MaskValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- ResolveScope ---

func TestResolveScope(t *testing.T) {
	t.Run("flag override", func(t *testing.T) {
		v := setupTestVault(t)
		if err := v.CreateProject("myapp"); err != nil {
			t.Fatal(err)
		}
		if err := v.CreateConfig("myapp", "dev"); err != nil {
			t.Fatal(err)
		}
		if err := v.CreateConfig("myapp", "prod"); err != nil {
			t.Fatal(err)
		}
		// Set active to dev
		if err := v.SetActiveScope("myapp", "dev"); err != nil {
			t.Fatal(err)
		}

		// Flag overrides active
		s, err := v.ResolveScope("myapp", "prod")
		if err != nil {
			t.Fatalf("ResolveScope() error: %v", err)
		}
		if s.Project != "myapp" || s.Config != "prod" {
			t.Errorf("ResolveScope() = %v, want myapp.prod", s)
		}
	})

	t.Run("active scope fallback", func(t *testing.T) {
		v := setupTestVault(t)
		if err := v.CreateProject("myapp"); err != nil {
			t.Fatal(err)
		}
		if err := v.CreateConfig("myapp", "dev"); err != nil {
			t.Fatal(err)
		}
		if err := v.SetActiveScope("myapp", "dev"); err != nil {
			t.Fatal(err)
		}

		s, err := v.ResolveScope("", "")
		if err != nil {
			t.Fatalf("ResolveScope() error: %v", err)
		}
		if s.Project != "myapp" || s.Config != "dev" {
			t.Errorf("ResolveScope() = %v, want myapp.dev", s)
		}
	})

	t.Run("missing project error", func(t *testing.T) {
		v := setupTestVault(t)
		// No projects, no active scope
		_, err := v.ResolveScope("", "")
		if err == nil {
			t.Fatal("ResolveScope() should error with no project")
		}
	})

	t.Run("missing config error", func(t *testing.T) {
		v := setupTestVault(t)
		if err := v.CreateProject("myapp"); err != nil {
			t.Fatal(err)
		}
		// Project exists but no config and no active config
		_, err := v.ResolveScope("myapp", "")
		if err == nil {
			t.Fatal("ResolveScope() should error with no config")
		}
	})
}

func TestScopeString(t *testing.T) {
	s := Scope{Project: "myapp", Config: "dev"}
	if got := s.String(); got != "myapp.dev" {
		t.Errorf("Scope.String() = %q, want %q", got, "myapp.dev")
	}
}
