package cmd

import (
	"fmt"

	"github.com/Spark-Rewards/homebrew-spk/internal/workspace"
	"github.com/spf13/cobra"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Show current workspace info",
	Long: `Displays information about the current workspace including name,
AWS profile, registered repos, and environment variables.

Example:
  spk workspace`,
	Aliases: []string{"ws"},
	RunE: func(cmd *cobra.Command, args []string) error {
		wsPath, err := workspace.Find()
		if err != nil {
			return err
		}

		ws, err := workspace.Load(wsPath)
		if err != nil {
			return err
		}

		fmt.Printf("Workspace:   %s\n", ws.Name)
		fmt.Printf("Location:    %s\n", wsPath)
		fmt.Printf("Created:     %s\n", ws.CreatedAt)
		fmt.Printf("AWS Profile: %s\n", orDefault(ws.AWSProfile, "(not set)"))
		fmt.Printf("AWS Region:  %s\n", orDefault(ws.AWSRegion, "(not set)"))
		fmt.Printf("Repos:       %d\n", len(ws.Repos))

		if len(ws.Env) > 0 {
			fmt.Println("\nEnvironment:")
			for k, v := range ws.Env {
				fmt.Printf("  %s=%s\n", k, v)
			}
		}

		if len(ws.Repos) > 0 {
			fmt.Println("\nRepositories:")
			for name, repo := range ws.Repos {
				deps := ""
				if len(repo.Dependencies) > 0 {
					deps = fmt.Sprintf(" (depends: %v)", repo.Dependencies)
				}
				fmt.Printf("  %s → %s%s\n", name, repo.Remote, deps)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
}
