package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version  = "dev"
	project  string
	config   string
	output   string
	noColor  bool
	verbose  bool

	rootCmd = &cobra.Command{
		Use:   "localvault",
		Short: "A local secrets manager backed by your OS keyring",
		Long: `localvault is a Doppler-like local secrets manager that stores
secrets encrypted in your operating system's keyring.

Manage projects, configs, and secrets without cloud dependencies.
Secrets never leave your machine unless you explicitly export them.`,
		Version:            version,
		SilenceUsage:       true,
		SilenceErrors:      true,
		DisableSuggestions: true,
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&project, "project", "p", "", "Project name")
	rootCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "Config name")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "Output format (table, json, dotenv)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
