package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Spark-Rewards/homebrew-spk/internal/git"
	"github.com/Spark-Rewards/homebrew-spk/internal/workspace"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all repos in the current workspace",
	Long: `Shows all repos registered in the workspace, their branch, and status.

Example:
  spk list`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		wsPath, err := workspace.Find()
		if err != nil {
			return err
		}

		ws, err := workspace.Load(wsPath)
		if err != nil {
			return err
		}

		if len(ws.Repos) == 0 {
			fmt.Println("No repos in workspace — run 'spk use <org/repo>' to add one")
			return nil
		}

		fmt.Printf("Workspace: %s (%s)\n\n", ws.Name, wsPath)
		fmt.Printf("%-25s %-15s %-10s %s\n", "REPO", "BRANCH", "STATUS", "PATH")
		fmt.Printf("%-25s %-15s %-10s %s\n", "----", "------", "------", "----")

		for name, repo := range ws.Repos {
			repoDir := filepath.Join(wsPath, repo.Path)

			branch := "-"
			status := "missing"

			if _, err := os.Stat(repoDir); err == nil {
				if git.IsRepo(repoDir) {
					b, err := git.CurrentBranch(repoDir)
					if err == nil {
						branch = b
					}
					s, err := git.Status(repoDir)
					if err == nil {
						if s == "" {
							status = "clean"
						} else {
							status = "dirty"
						}
					}
				} else {
					status = "not-git"
				}
			}

			fmt.Printf("%-25s %-15s %-10s %s\n", name, branch, status, repo.Path)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
