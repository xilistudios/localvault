package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xilistudios/localvault/internal/vault"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List all projects",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}

		projects := v.ListProjects()
		if len(projects) == 0 {
			fmt.Println("No projects. Create one with: localvault projects create <name>")
			return nil
		}

		for _, name := range projects {
			p := v.Meta.Projects[name]
			marker := "  "
			if name == v.Meta.ActiveProject {
				marker = "* "
			}
			fmt.Printf("%s%s (%d configs)\n", marker, name, len(p.Configs))
		}
		return nil
	},
}

var projectsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}
		if err := v.CreateProject(args[0]); err != nil {
			return err
		}
		fmt.Printf("✓ Created project %q\n", args[0])
		return nil
	},
}

var projectsDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a project and all its configs/secrets",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}
		if err := v.DeleteProject(args[0]); err != nil {
			return err
		}
		fmt.Printf("✓ Deleted project %q\n", args[0])
		return nil
	},
}

var projectsInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show project details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.NewVault(nil)
		if err != nil {
			return err
		}
		p, err := v.Meta.GetProject(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Project:   %s\n", args[0])
		fmt.Printf("Created:   %s\n", p.CreatedAt.Format("2006-01-02 15:04:05 UTC"))
		fmt.Printf("Configs:   %d\n", len(p.Configs))
		for name, cfg := range p.Configs {
			fmt.Printf("  - %s (%d secrets)\n", name, len(cfg.Secrets))
		}
		return nil
	},
}

func init() {
	projectsCmd.AddCommand(projectsCreateCmd)
	projectsCmd.AddCommand(projectsDeleteCmd)
	projectsCmd.AddCommand(projectsInfoCmd)
	rootCmd.AddCommand(projectsCmd)
}
