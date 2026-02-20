package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Spark-Rewards/homebrew-spark-cli/internal/git"
	"github.com/Spark-Rewards/homebrew-spark-cli/internal/workspace"
	"github.com/spf13/cobra"
)

var workspaceCmd = &cobra.Command{
	Use:     "info",
	Short:   "Show workspace info and repo status",
	Aliases: []string{"workspace", "ws", "status"},
	Long: `Displays workspace info including repos, their git status, and environment.

Examples:
  spark-cli info
  spark-cli status
  spark-cli ws`,
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
		fmt.Printf("AWS Profile: %s\n", orDefault(ws.AWSProfile, "(not set)"))
		fmt.Printf("Environment: %s\n", orDefault(ws.SSMEnvPath, "beta"))
		fmt.Println()

		if len(ws.Repos) > 0 {
			fmt.Printf("%-20s %-15s %-10s %s\n", "REPO", "BRANCH", "STATUS", "PATH")
			fmt.Printf("%-20s %-15s %-10s %s\n", "----", "------", "------", "----")

			for name, repo := range ws.Repos {
				repoDir := filepath.Join(wsPath, repo.Path)
				branch := "-"
				status := "missing"

				if _, err := os.Stat(repoDir); err == nil {
					if git.IsRepo(repoDir) {
						b, _ := git.CurrentBranch(repoDir)
						if b != "" {
							branch = b
						}
						if git.IsDirty(repoDir) {
							status = "dirty"
						} else {
							status = "clean"
						}
					}
				}

				fmt.Printf("%-20s %-15s %-10s %s\n", name, branch, status, repo.Path)
			}
		} else {
			fmt.Println("No repos — run 'spark-cli use <repo>' to add one")
		}

		globalEnv, _ := workspace.ReadGlobalEnv(wsPath)
		if len(globalEnv) > 0 {
			fmt.Println("\nEnvironment (.env):")
			for k, v := range globalEnv {
				display := v
				if len(display) > 40 {
					display = display[:37] + "..."
				}
				fmt.Printf("  %s=%s\n", k, display)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(workspaceCmd)
}
