package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/xilistudios/localvault/internal/model"
)

const (
	vaultFileName = "vault.json"
	lockFileName  = "vault.lock"
)

// VaultFilePath returns ~/.localvault/vault.json
func VaultFilePath() (string, error) {
	dir, err := VaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, vaultFileName), nil
}

// IsInitialized checks if the vault directory and file exist
func IsInitialized() bool {
	path, err := VaultFilePath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// InitVault creates ~/.localvault/ and an empty vault.json
func InitVault() error {
	dir, err := VaultDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("cannot create vault directory: %w", err)
	}

	vf := model.NewVaultFile()
	return writeMetadata(vf)
}

// LoadMetadata reads and parses vault.json
func LoadMetadata() (*model.VaultFile, error) {
	path, err := VaultFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read vault file: %w", err)
	}
	var vf model.VaultFile
	if err := json.Unmarshal(data, &vf); err != nil {
		return nil, fmt.Errorf("cannot parse vault file: %w", err)
	}
	return &vf, nil
}

// SaveMetadata writes vault.json with file locking
func SaveMetadata(vf *model.VaultFile) error {
	return writeMetadata(vf)
}

func writeMetadata(vf *model.VaultFile) error {
	path, err := VaultFilePath()
	if err != nil {
		return err
	}

	// Acquire file lock
	lockPath := filepath.Join(filepath.Dir(path), lockFileName)
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("cannot open lock file: %w", err)
	}
	defer lockFile.Close()

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("cannot acquire lock: %w", err)
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)

	data, err := json.MarshalIndent(vf, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot serialize vault: %w", err)
	}

	// Write atomically: temp file + rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("cannot write vault file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cannot finalize vault file: %w", err)
	}
	return nil
}
