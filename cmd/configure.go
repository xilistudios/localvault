package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xilistudios/localvault/internal/vault"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Set the active project and config",
	Long: `Set the active project and/or config scope.
Use --project and --config flags, or run interactively.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !vault.IsInitialized() {
			return fmt.Errorf("vault not initialized: run 'localvault setup' first")
		}

		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}

		p := project // from --project flag
		c := config  // from --config flag

		if p == "" && c == "" {
			// Interactive mode: show current and prompt
			fmt.Printf("Current scope: %s\n", scopeDisplay(v))
			fmt.Println("Use --project and --config flags to set scope.")
			fmt.Println("Example: localvault configure --project myapp --config dev")
			return nil
		}

		if err := v.SetActiveScope(p, c); err != nil {
			return err
		}

		// Reload to show updated state
		v, _ = vault.NewVault(nil)
		fmt.Printf("✓ Active scope: %s\n", scopeDisplay(v))
		return nil
	},
}

func scopeDisplay(v *vault.Vault) string {
	if v.Meta.ActiveProject == "" {
		return "(none)"
	}
	s := v.Meta.ActiveProject
	if v.Meta.ActiveConfig != "" {
		s += "." + v.Meta.ActiveConfig
	}
	return s
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
