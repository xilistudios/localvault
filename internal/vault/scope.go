package vault

import (
	"fmt"

	"github.com/xilistudios/localvault/internal/model"
)

// Scope represents a resolved project+config pair
type Scope struct {
	Project string
	Config  string
}

func (s Scope) String() string {
	return s.Project + "." + s.Config
}

// ResolveScope determines the effective project+config.
// Priority: explicit flags > active scope from metadata.
// Returns error if scope cannot be determined.
func ResolveScope(vf *model.VaultFile, flagProject, flagConfig string) (Scope, error) {
	p := flagProject
	c := flagConfig

	if p == "" {
		p = vf.ActiveProject
	}
	if c == "" {
		c = vf.ActiveConfig
	}

	if p == "" {
		return Scope{}, fmt.Errorf("no project specified: use --project flag or run 'localvault configure'")
	}
	if c == "" {
		return Scope{}, fmt.Errorf("no config specified: use --config flag or run 'localvault configure'")
	}

	// Validate they exist
	if _, err := vf.GetProject(p); err != nil {
		return Scope{}, err
	}
	if _, err := vf.GetConfig(p, c); err != nil {
		return Scope{}, err
	}

	return Scope{Project: p, Config: c}, nil
}
