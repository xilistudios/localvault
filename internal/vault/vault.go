package vault

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xilistudios/localvault/internal/model"
)

// Vault orchestrates metadata and secret storage
type Vault struct {
	Meta  *model.VaultFile
	Store SecretStore
}

// NewVault creates a Vault with the given secret store and loads metadata.
// If store is nil, it creates an OSKeyring.
func NewVault(store SecretStore) (*Vault, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("vault not initialized: run 'localvault setup' first")
	}

	meta, err := LoadMetadata()
	if err != nil {
		return nil, err
	}

	if store == nil {
		store, err = NewOSKeyring()
		if err != nil {
			return nil, err
		}
	}

	return &Vault{Meta: meta, Store: store}, nil
}

// Save persists metadata changes to disk
func (v *Vault) Save() error {
	return SaveMetadata(v.Meta)
}

// ResolveScope resolves the effective project+config from flags or active scope
func (v *Vault) ResolveScope(flagProject, flagConfig string) (Scope, error) {
	return ResolveScope(v.Meta, flagProject, flagConfig)
}

// --- Project operations ---

func (v *Vault) ListProjects() []string {
	names := make([]string, 0, len(v.Meta.Projects))
	for name := range v.Meta.Projects {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (v *Vault) CreateProject(name string) error {
	if err := v.Meta.AddProject(name); err != nil {
		return err
	}
	return v.Save()
}

func (v *Vault) DeleteProject(name string) error {
	p, err := v.Meta.GetProject(name)
	if err != nil {
		return err
	}

	// Delete all secrets for all configs in this project
	for configName, cfg := range p.Configs {
		for _, secretName := range cfg.Secrets {
			key := SecretKey(name, configName, secretName)
			if err := v.Store.Remove(key); err != nil {
				return fmt.Errorf("failed to delete secret %s: %w", key, err)
			}
		}
	}

	if err := v.Meta.DeleteProject(name); err != nil {
		return err
	}
	return v.Save()
}

// --- Config operations ---

func (v *Vault) ListConfigs(project string) ([]string, error) {
	p, err := v.Meta.GetProject(project)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(p.Configs))
	for name := range p.Configs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func (v *Vault) CreateConfig(project, name string) error {
	if err := v.Meta.AddConfig(project, name); err != nil {
		return err
	}
	return v.Save()
}

func (v *Vault) DeleteConfig(project, name string) error {
	cfg, err := v.Meta.GetConfig(project, name)
	if err != nil {
		return err
	}

	// Delete all secrets for this config
	for _, secretName := range cfg.Secrets {
		key := SecretKey(project, name, secretName)
		if err := v.Store.Remove(key); err != nil {
			return fmt.Errorf("failed to delete secret %s: %w", key, err)
		}
	}

	if err := v.Meta.DeleteConfig(project, name); err != nil {
		return err
	}
	return v.Save()
}

func (v *Vault) CopyConfig(project, src, dst string) error {
	srcCfg, err := v.Meta.GetConfig(project, src)
	if err != nil {
		return err
	}

	// Create dst config if it doesn't exist
	if _, err := v.Meta.GetConfig(project, dst); err != nil {
		if err := v.Meta.AddConfig(project, dst); err != nil {
			return err
		}
	}

	// Copy each secret
	for _, secretName := range srcCfg.Secrets {
		srcKey := SecretKey(project, src, secretName)
		data, err := v.Store.Get(srcKey)
		if err != nil {
			return fmt.Errorf("failed to read secret %s: %w", secretName, err)
		}
		dstKey := SecretKey(project, dst, secretName)
		if err := v.Store.Set(dstKey, data); err != nil {
			return fmt.Errorf("failed to write secret %s: %w", secretName, err)
		}
		if err := v.Meta.AddSecretToIndex(project, dst, secretName); err != nil {
			return err
		}
	}

	return v.Save()
}

// --- Scope operations ---

func (v *Vault) SetActiveScope(project, config string) error {
	if project != "" {
		if _, err := v.Meta.GetProject(project); err != nil {
			return err
		}
		v.Meta.ActiveProject = project
	}
	if config != "" {
		p := v.Meta.ActiveProject
		if p == "" {
			return fmt.Errorf("no active project: set --project first")
		}
		if _, err := v.Meta.GetConfig(p, config); err != nil {
			return err
		}
		v.Meta.ActiveConfig = config
	}
	return v.Save()
}

// --- Secret operations ---

func (v *Vault) SetSecret(project, config, key, value string) error {
	if err := model.ValidateSecretKey(key); err != nil {
		return err
	}
	if err := model.ValidateSecretValue(value); err != nil {
		return err
	}

	k := SecretKey(project, config, key)
	if err := v.Store.Set(k, []byte(value)); err != nil {
		return err
	}
	if err := v.Meta.AddSecretToIndex(project, config, key); err != nil {
		return err
	}
	return v.Save()
}

func (v *Vault) GetSecret(project, config, key string) (string, error) {
	k := SecretKey(project, config, key)
	data, err := v.Store.Get(k)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (v *Vault) UnsetSecret(project, config, key string) error {
	k := SecretKey(project, config, key)
	if err := v.Store.Remove(k); err != nil {
		return err
	}
	if err := v.Meta.RemoveSecretFromIndex(project, config, key); err != nil {
		return err
	}
	return v.Save()
}

// ListSecrets returns all secret key names for a scope (from metadata index)
func (v *Vault) ListSecrets(project, config string) ([]string, error) {
	cfg, err := v.Meta.GetConfig(project, config)
	if err != nil {
		return nil, err
	}
	secrets := make([]string, len(cfg.Secrets))
	copy(secrets, cfg.Secrets)
	sort.Strings(secrets)
	return secrets, nil
}

// GetAllSecrets returns all key-value pairs for a scope
func (v *Vault) GetAllSecrets(project, config string) (map[string]string, error) {
	names, err := v.ListSecrets(project, config)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(names))
	for _, name := range names {
		val, err := v.GetSecret(project, config, name)
		if err != nil {
			return nil, fmt.Errorf("failed to read secret %s: %w", name, err)
		}
		result[name] = val
	}
	return result, nil
}

// MaskValue masks a secret value for display: shows first 2 and last 2 chars
func MaskValue(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}
