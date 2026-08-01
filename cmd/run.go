package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/xilistudios/localvault/internal/vault"
)

var runCmd = &cobra.Command{
	Use:   "run -- <command> [args...]",
	Short: "Run a command with secrets injected as environment variables",
	Long: `Run a command with all secrets from the active scope injected as env vars.

Examples:
  localvault run -- node server.js
  localvault run --project myapp --config prd -- ./deploy.sh
  localvault run -- env | grep DATABASE`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: false,
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

		// Build environment: current env + secrets
		env := os.Environ()
		for k, val := range secrets {
			env = append(env, fmt.Sprintf("%s=%s", k, val))
		}

		// Find the command
		binary, err := exec.LookPath(args[0])
		if err != nil {
			return fmt.Errorf("command not found: %s", args[0])
		}

		// Exec (replaces current process)
		return syscall.Exec(binary, args, env)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
