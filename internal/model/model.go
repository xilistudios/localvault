package model

import (
	"fmt"
	"regexp"
	"time"
)

// VaultFile is the top-level structure stored in ~/.localvault/vault.json
type VaultFile struct {
	Version        int                 `json:"version"`
	ActiveProject  string              `json:"active_project,omitempty"`
	ActiveConfig   string              `json:"active_config,omitempty"`
	Projects       map[string]*Project `json:"projects"`
}

type Project struct {
	CreatedAt time.Time           `json:"created_at"`
	Configs   map[string]*Config `json:"configs"`
}

type Config struct {
	CreatedAt time.Time `json:"created_at"`
	Secrets   []string  `json:"secrets"` // key names only, no values
}

// Validation
var (
	nameRegex   = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)
	secretRegex = regexp.MustCompile(`^[A-Z0-9_]{1,256}$`)
)

func ValidateProjectName(name string) error {
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("invalid project name %q: must match [a-z0-9_-], 1-64 chars", name)
	}
	return nil
}

func ValidateConfigName(name string) error {
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("invalid config name %q: must match [a-z0-9_-], 1-64 chars", name)
	}
	return nil
}

func ValidateSecretKey(key string) error {
	if !secretRegex.MatchString(key) {
		return fmt.Errorf("invalid secret key %q: must match [A-Z0-9_], 1-256 chars", key)
	}
	return nil
}

func ValidateSecretValue(value string) error {
	if len(value) > 64*1024 {
		return fmt.Errorf("secret value exceeds 64KB limit")
	}
	return nil
}

// NewVaultFile returns an initialized empty vault
func NewVaultFile() *VaultFile {
	return &VaultFile{
		Version:  1,
		Projects: make(map[string]*Project),
	}
}

// Helper methods on VaultFile

func (v *VaultFile) GetProject(name string) (*Project, error) {
	p, ok := v.Projects[name]
	if !ok {
		return nil, fmt.Errorf("project %q not found", name)
	}
	return p, nil
}

func (v *VaultFile) AddProject(name string) error {
	if err := ValidateProjectName(name); err != nil {
		return err
	}
	if _, exists := v.Projects[name]; exists {
		return fmt.Errorf("project %q already exists", name)
	}
	v.Projects[name] = &Project{
		CreatedAt: time.Now().UTC(),
		Configs:   make(map[string]*Config),
	}
	return nil
}

func (v *VaultFile) DeleteProject(name string) error {
	if _, err := v.GetProject(name); err != nil {
		return err
	}
	delete(v.Projects, name)
	if v.ActiveProject == name {
		v.ActiveProject = ""
		v.ActiveConfig = ""
	}
	return nil
}

func (v *VaultFile) AddConfig(project, config string) error {
	if err := ValidateConfigName(config); err != nil {
		return err
	}
	p, err := v.GetProject(project)
	if err != nil {
		return err
	}
	if _, exists := p.Configs[config]; exists {
		return fmt.Errorf("config %q already exists in project %q", config, project)
	}
	p.Configs[config] = &Config{
		CreatedAt: time.Now().UTC(),
		Secrets:   []string{},
	}
	return nil
}

func (v *VaultFile) DeleteConfig(project, config string) error {
	p, err := v.GetProject(project)
	if err != nil {
		return err
	}
	if _, exists := p.Configs[config]; !exists {
		return fmt.Errorf("config %q not found in project %q", config, project)
	}
	delete(p.Configs, config)
	if v.ActiveProject == project && v.ActiveConfig == config {
		v.ActiveConfig = ""
	}
	return nil
}

func (v *VaultFile) GetConfig(project, config string) (*Config, error) {
	p, err := v.GetProject(project)
	if err != nil {
		return nil, err
	}
	c, ok := p.Configs[config]
	if !ok {
		return nil, fmt.Errorf("config %q not found in project %q", config, project)
	}
	return c, nil
}

// AddSecretToIndex adds a secret key name to the config's index (idempotent)
func (v *VaultFile) AddSecretToIndex(project, config, key string) error {
	c, err := v.GetConfig(project, config)
	if err != nil {
		return err
	}
	for _, s := range c.Secrets {
		if s == key {
			return nil // already indexed
		}
	}
	c.Secrets = append(c.Secrets, key)
	return nil
}

// RemoveSecretFromIndex removes a secret key name from the config's index
func (v *VaultFile) RemoveSecretFromIndex(project, config, key string) error {
	c, err := v.GetConfig(project, config)
	if err != nil {
		return err
	}
	for i, s := range c.Secrets {
		if s == key {
			c.Secrets = append(c.Secrets[:i], c.Secrets[i+1:]...)
			return nil
		}
	}
	return nil // not found, no-op
}
