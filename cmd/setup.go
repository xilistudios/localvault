package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xilistudios/localvault/internal/vault"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Initialize the localvault directory",
	Long:  "Creates ~/.localvault/ with an empty vault.json. Run this once before using other commands.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if vault.IsInitialized() {
			dir, _ := vault.VaultDir()
			fmt.Printf("Vault already initialized at %s\n", dir)
			return nil
		}
		if err := vault.InitVault(); err != nil {
			return err
		}
		dir, _ := vault.VaultDir()
		fmt.Printf("✓ Vault initialized at %s\n", dir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
