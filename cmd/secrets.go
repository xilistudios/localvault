package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xilistudios/localvault/internal/vault"
	"golang.org/x/term"
)

// localvault set KEY=VALUE [KEY2=VALUE2 ...]
// localvault set KEY  (prompts for hidden input)
var setCmd = &cobra.Command{
	Use:   "set KEY=VALUE [KEY=VALUE...]",
	Short: "Set one or more secrets",
	Long: `Set secrets in the active scope.

Examples:
  localvault set DATABASE_URL=postgres://localhost/mydb
  localvault set API_KEY=abc123 SECRET_TOKEN=xyz
  localvault set PASSWORD          (prompts for hidden input)`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}
		scope, err := v.ResolveScope(project, config)
		if err != nil {
			return err
		}

		count := 0
		for _, arg := range args {
			var key, value string
			if idx := strings.Index(arg, "="); idx > 0 {
				key = arg[:idx]
				value = arg[idx+1:]
			} else {
				// No = sign, prompt for value
				key = arg
				fmt.Fprintf(os.Stderr, "? Enter value for %s: ", key)
				password, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}
				fmt.Fprintln(os.Stderr)
				value = string(password)
			}

			if err := v.SetSecret(scope.Project, scope.Config, key, value); err != nil {
				return err
			}
			count++
		}

		if count == 1 {
			fmt.Printf("✓ Set %s in %s\n", args[0], scope.String())
		} else {
			fmt.Printf("✓ Set %d secrets in %s\n", count, scope.String())
		}
		return nil
	},
}

// localvault get KEY [--plain]
var getPlain bool

var getCmd = &cobra.Command{
	Use:   "get <KEY>",
	Short: "Get a secret value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}
		scope, err := v.ResolveScope(project, config)
		if err != nil {
			return err
		}

		value, err := v.GetSecret(scope.Project, scope.Config, args[0])
		if err != nil {
			return err
		}

		if getPlain {
			fmt.Print(value)
		} else {
			fmt.Printf("%s=%s\n", args[0], value)
		}
		return nil
	},
}

// localvault secrets [--values]
var showValues bool

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "List all secrets in the active scope",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}
		scope, err := v.ResolveScope(project, config)
		if err != nil {
			return err
		}

		names, err := v.ListSecrets(scope.Project, scope.Config)
		if err != nil {
			return err
		}

		if len(names) == 0 {
			fmt.Printf("No secrets in %s\n", scope.String())
			return nil
		}

		// Find max key length for alignment
		maxLen := 0
		for _, name := range names {
			if len(name) > maxLen {
				maxLen = len(name)
			}
		}

		for _, name := range names {
			if showValues {
				value, err := v.GetSecret(scope.Project, scope.Config, name)
				if err != nil {
					return err
				}
				fmt.Printf("  %-*s = %s\n", maxLen, name, value)
			} else {
				value, err := v.GetSecret(scope.Project, scope.Config, name)
				if err != nil {
					// If we can't read the value, just show the key
					fmt.Printf("  %-*s = (error reading)\n", maxLen, name)
					continue
				}
				fmt.Printf("  %-*s = %s\n", maxLen, name, vault.MaskValue(value))
			}
		}
		return nil
	},
}

// localvault unset KEY [KEY2 ...]
var unsetCmd = &cobra.Command{
	Use:   "unset KEY [KEY...]",
	Short: "Delete one or more secrets",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}
		scope, err := v.ResolveScope(project, config)
		if err != nil {
			return err
		}

		for _, key := range args {
			if err := v.UnsetSecret(scope.Project, scope.Config, key); err != nil {
				return err
			}
		}

		if len(args) == 1 {
			fmt.Printf("✓ Removed %s from %s\n", args[0], scope.String())
		} else {
			fmt.Printf("✓ Removed %d secrets from %s\n", len(args), scope.String())
		}
		return nil
	},
}

func init() {
	getCmd.Flags().BoolVar(&getPlain, "plain", false, "Output raw value (no key= prefix)")
	secretsCmd.Flags().BoolVar(&showValues, "values", false, "Show actual values instead of masked")

	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(secretsCmd)
	rootCmd.AddCommand(unsetCmd)
}
