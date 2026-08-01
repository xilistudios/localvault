package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xilistudios/localvault/internal/envfile"
	"github.com/xilistudios/localvault/internal/vault"
)

var importStdin bool

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import secrets from a .env file",
	Long: `Import secrets from a .env file or stdin into the active scope.

Examples:
  localvault import .env
  localvault import .env.production
  cat secrets.env | localvault import --stdin`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}
		scope, err := v.ResolveScope(project, config)
		if err != nil {
			return err
		}

		var reader *os.File
		if importStdin {
			reader = os.Stdin
		} else if len(args) == 1 {
			reader, err = os.Open(args[0])
			if err != nil {
				return fmt.Errorf("cannot open file: %w", err)
			}
			defer reader.Close()
		} else {
			return fmt.Errorf("specify a file or use --stdin")
		}

		pairs, err := envfile.Parse(reader)
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}

		if len(pairs) == 0 {
			fmt.Println("No secrets found in input.")
			return nil
		}

		imported := 0
		for _, p := range pairs {
			if err := v.SetSecret(scope.Project, scope.Config, p.Key, p.Value); err != nil {
				return fmt.Errorf("failed to set %s: %w", p.Key, err)
			}
			imported++
		}

		fmt.Printf("Imported %d secret(s) into %s\n", imported, scope)
		return nil
	},
}

var exportFormat string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export secrets from the active scope",
	Long: `Export all secrets from the active scope in the specified format.

Supported formats: dotenv (default), json, docker

Examples:
  localvault export
  localvault export --format json
  localvault export --format dotenv > .env
  localvault export --format docker`,
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}
		scope, err := v.ResolveScope(project, config)
		if err != nil {
			return err
		}

		secrets, err := v.GetAllSecrets(scope.Project, scope.Config)
		if err != nil {
			return err
		}

		if len(secrets) == 0 {
			fmt.Println("No secrets found in scope.")
			return nil
		}

		format := envfile.Format(strings.ToLower(exportFormat))
		out, err := envfile.Export(secrets, format)
		if err != nil {
			return err
		}

		fmt.Print(out)
		return nil
	},
}

func init() {
	importCmd.Flags().BoolVar(&importStdin, "stdin", false, "Read from stdin instead of a file")
	rootCmd.AddCommand(importCmd)

	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "dotenv", "Export format: dotenv, json, docker")
	rootCmd.AddCommand(exportCmd)
}
