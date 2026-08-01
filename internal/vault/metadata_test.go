package vault

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/xilistudios/localvault/internal/model"
)

func TestVaultDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	dir, err := VaultDir()
	if err != nil {
		t.Fatalf("VaultDir() error: %v", err)
	}
	if !strings.HasSuffix(dir, ".localvault") {
		t.Errorf("expected path ending in .localvault, got %s", dir)
	}
}

func TestVaultFilePath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	path, err := VaultFilePath()
	if err != nil {
		t.Fatalf("VaultFilePath() error: %v", err)
	}
	if !strings.HasSuffix(path, "vault.json") {
		t.Errorf("expected path ending in vault.json, got %s", path)
	}
}

func TestIsInitialized(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if IsInitialized() {
		t.Error("IsInitialized should be false before InitVault")
	}

	if err := InitVault(); err != nil {
		t.Fatalf("InitVault() error: %v", err)
	}

	if !IsInitialized() {
		t.Error("IsInitialized should be true after InitVault")
	}
}

func TestInitVault(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := InitVault(); err != nil {
		t.Fatalf("InitVault() error: %v", err)
	}

	dir, _ := VaultDir()

	// Check directory exists
	if _, err := filepath.Abs(dir); err != nil {
		t.Fatalf("filepath.Abs error: %v", err)
	}

	// Check file exists by loading
	vf, err := LoadMetadata()
	if err != nil {
		t.Fatalf("LoadMetadata() after InitVault error: %v", err)
	}
	if vf.Version != 1 {
		t.Errorf("expected version 1, got %d", vf.Version)
	}
	if vf.Projects == nil {
		t.Error("Projects map should not be nil")
	}
}

func TestLoadMetadata(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := InitVault(); err != nil {
		t.Fatalf("InitVault() error: %v", err)
	}

	vf, err := LoadMetadata()
	if err != nil {
		t.Fatalf("LoadMetadata() error: %v", err)
	}
	if vf == nil {
		t.Fatal("LoadMetadata returned nil")
	}
	if vf.Version != 1 {
		t.Errorf("expected version 1, got %d", vf.Version)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := InitVault(); err != nil {
		t.Fatalf("InitVault() error: %v", err)
	}

	// Load, modify, save
	vf, err := LoadMetadata()
	if err != nil {
		t.Fatalf("LoadMetadata() error: %v", err)
	}

	if err := vf.AddProject("testproject"); err != nil {
		t.Fatalf("AddProject error: %v", err)
	}
	if err := vf.AddConfig("testproject", "staging"); err != nil {
		t.Fatalf("AddConfig error: %v", err)
	}

	if err := SaveMetadata(vf); err != nil {
		t.Fatalf("SaveMetadata() error: %v", err)
	}

	// Reload and verify
	vf2, err := LoadMetadata()
	if err != nil {
		t.Fatalf("LoadMetadata() after save error: %v", err)
	}

	proj, err := vf2.GetProject("testproject")
	if err != nil {
		t.Fatalf("project not found after round-trip: %v", err)
	}
	if _, ok := proj.Configs["staging"]; !ok {
		t.Error("config 'staging' not found after round-trip")
	}
}

func TestConcurrentSave(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := InitVault(); err != nil {
		t.Fatalf("InitVault() error: %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			vf := model.NewVaultFile()
			if err := SaveMetadata(vf); err != nil {
				errs <- err
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent save error: %v", err)
	}

	// Verify the file is still valid
	vf, err := LoadMetadata()
	if err != nil {
		t.Fatalf("LoadMetadata() after concurrent saves error: %v", err)
	}
	if vf.Version != 1 {
		t.Errorf("expected version 1 after concurrent saves, got %d", vf.Version)
	}
}
