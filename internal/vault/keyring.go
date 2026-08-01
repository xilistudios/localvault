package vault

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/99designs/keyring"
)

const (
	keyringServiceName = "localvault"
	keyringPrefix      = "localvault."
)

// SecretStore is the interface for secret storage backends
type SecretStore interface {
	Get(key string) ([]byte, error)
	Set(key string, data []byte) error
	Remove(key string) error
	Keys() ([]string, error)
}

// VaultDir returns the localvault configuration directory.
// Uses LOCALVAULT_DIR env var, or defaults to ~/.localvault.
func VaultDir() (string, error) {
	if dir := os.Getenv("LOCALVAULT_DIR"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".localvault"), nil
}

// OSKeyring implements SecretStore using 99designs/keyring
type OSKeyring struct {
	ring keyring.Keyring
}

// NewOSKeyring creates a keyring backend. It tries OS keyring first,
// falls back to encrypted file for headless/CI environments.
func NewOSKeyring() (*OSKeyring, error) {
	dir, err := VaultDir()
	if err != nil {
		return nil, err
	}

	// Try to open with allowed backends
	ring, err := keyring.Open(keyring.Config{
		ServiceName: keyringServiceName,
		AllowedBackends: []keyring.BackendType{
			keyring.SecretServiceBackend,
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.KWalletBackend,
			keyring.FileBackend,
		},
		FileDir:          filepath.Join(dir, "keyring"),
		FilePasswordFunc: fileKeyringPassphrase,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot open keyring: %w", err)
	}

	return &OSKeyring{ring: ring}, nil
}

// fileKeyringPassphrase provides the passphrase for the file backend.
// Uses LOCALVAULT_PASSPHRASE env var, or defaults to "localvault" for convenience.
func fileKeyringPassphrase(prompt string) (string, error) {
	if pass := os.Getenv("LOCALVAULT_PASSPHRASE"); pass != "" {
		return pass, nil
	}
	// Default passphrase for local dev convenience.
	// In production CI, set LOCALVAULT_PASSPHRASE.
	return "localvault", nil
}

func (k *OSKeyring) Get(key string) ([]byte, error) {
	item, err := k.ring.Get(key)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return nil, fmt.Errorf("secret %q not found", key)
		}
		return nil, fmt.Errorf("keyring get error: %w", err)
	}
	return item.Data, nil
}

func (k *OSKeyring) Set(key string, data []byte) error {
	item := keyring.Item{
		Key:         key,
		Data:        data,
		Label:       fmt.Sprintf("localvault: %s", key),
		Description: "Managed by localvault",
	}
	if err := k.ring.Set(item); err != nil {
		return fmt.Errorf("keyring set error: %w", err)
	}
	return nil
}

func (k *OSKeyring) Remove(key string) error {
	if err := k.ring.Remove(key); err != nil {
		if err == keyring.ErrKeyNotFound {
			return nil // idempotent
		}
		return fmt.Errorf("keyring remove error: %w", err)
	}
	return nil
}

func (k *OSKeyring) Keys() ([]string, error) {
	keys, err := k.ring.Keys()
	if err != nil {
		return nil, fmt.Errorf("keyring keys error: %w", err)
	}
	return keys, nil
}

// SecretKey builds the keyring key for a secret
func SecretKey(project, config, name string) string {
	return fmt.Sprintf("%s%s.%s.%s", keyringPrefix, project, config, name)
}

// SecretKeyPrefix builds the prefix for all secrets in a scope
func SecretKeyPrefix(project, config string) string {
	return fmt.Sprintf("%s%s.%s.", keyringPrefix, project, config)
}
