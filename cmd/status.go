package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xilistudios/localvault/internal/vault"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show vault status and active scope",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !vault.IsInitialized() {
			return fmt.Errorf("vault not initialized: run 'localvault setup' first")
		}

		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}

		dir, _ := vault.VaultDir()
		fmt.Printf("Vault:    %s\n", dir)
		fmt.Printf("Projects: %d\n", len(v.Meta.Projects))

		if v.Meta.ActiveProject != "" {
			fmt.Printf("Active:   %s", v.Meta.ActiveProject)
			if v.Meta.ActiveConfig != "" {
				fmt.Printf(".%s", v.Meta.ActiveConfig)
			}
			fmt.Println()
		} else {
			fmt.Println("Active:   (none)")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
