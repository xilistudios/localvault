package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xilistudios/localvault/internal/vault"
)

var configsCmd = &cobra.Command{
	Use:   "configs",
	Short: "List configs for the active project",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}

		proj := project
		if proj == "" {
			proj = v.Meta.ActiveProject
		}
		if proj == "" {
			return fmt.Errorf("no project specified: use --project flag or run 'localvault configure'")
		}

		configs, err := v.ListConfigs(proj)
		if err != nil {
			return err
		}
		if len(configs) == 0 {
			fmt.Printf("No configs in project %q. Create one with: localvault configs create <name>\n", proj)
			return nil
		}

		for _, name := range configs {
			marker := "  "
			if proj == v.Meta.ActiveProject && name == v.Meta.ActiveConfig {
				marker = "* "
			}
			cfg := v.Meta.Projects[proj].Configs[name]
			fmt.Printf("%s%s (%d secrets)\n", marker, name, len(cfg.Secrets))
		}
		return nil
	},
}

var configsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a config in the active project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}

		proj := project
		if proj == "" {
			proj = v.Meta.ActiveProject
		}
		if proj == "" {
			return fmt.Errorf("no project specified: use --project flag or run 'localvault configure'")
		}

		if err := v.CreateConfig(proj, args[0]); err != nil {
			return err
		}
		fmt.Printf("✓ Created config %q in project %q\n", args[0], proj)
		return nil
	},
}

var configsDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a config and its secrets",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}

		proj := project
		if proj == "" {
			proj = v.Meta.ActiveProject
		}
		if proj == "" {
			return fmt.Errorf("no project specified: use --project flag or run 'localvault configure'")
		}

		if err := v.DeleteConfig(proj, args[0]); err != nil {
			return err
		}
		fmt.Printf("✓ Deleted config %q from project %q\n", args[0], proj)
		return nil
	},
}

var configsCopyCmd = &cobra.Command{
	Use:   "copy <source> <destination>",
	Short: "Copy all secrets from one config to another",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}

		proj := project
		if proj == "" {
			proj = v.Meta.ActiveProject
		}
		if proj == "" {
			return fmt.Errorf("no project specified: use --project flag or run 'localvault configure'")
		}

		if err := v.CopyConfig(proj, args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("✓ Copied secrets from %s.%s → %s.%s\n", proj, args[0], proj, args[1])
		return nil
	},
}

func init() {
	configsCmd.AddCommand(configsCreateCmd)
	configsCmd.AddCommand(configsDeleteCmd)
	configsCmd.AddCommand(configsCopyCmd)
	rootCmd.AddCommand(configsCmd)
}
