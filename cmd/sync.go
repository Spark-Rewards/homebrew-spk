package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Spark-Rewards/homebrew-spk/internal/git"
	"github.com/Spark-Rewards/homebrew-spk/internal/workspace"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [repo-name]",
	Short: "Pull latest changes for workspace repos",
	Long: `Runs 'git pull' for all repos in the workspace, or a specific repo
if a name is provided.

Examples:
  spk sync                # pull all repos
  spk sync BusinessAPI    # pull specific repo`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wsPath, err := workspace.Find()
		if err != nil {
			return err
		}

		ws, err := workspace.Load(wsPath)
		if err != nil {
			return err
		}

		if len(args) == 1 {
			return syncRepo(wsPath, ws, args[0])
		}

		return syncAll(wsPath, ws)
	},
}

func syncRepo(wsPath string, ws *workspace.Workspace, name string) error {
	repo, ok := ws.Repos[name]
	if !ok {
		return fmt.Errorf("repo '%s' not found in workspace — run 'spk list' to see registered repos", name)
	}

	repoDir := filepath.Join(wsPath, repo.Path)
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("repo directory %s does not exist — run 'spk use' to clone it", repoDir)
	}

	fmt.Printf("Syncing %s...\n", name)
	branch, _ := git.CurrentBranch(repoDir)
	if branch != "" {
		fmt.Printf("  Branch: %s\n", branch)
	}

	if err := git.Pull(repoDir); err != nil {
		return fmt.Errorf("failed to sync %s: %w", name, err)
	}

	fmt.Printf("  %s synced\n", name)
	return nil
}

func syncAll(wsPath string, ws *workspace.Workspace) error {
	if len(ws.Repos) == 0 {
		fmt.Println("No repos in workspace — run 'spk use <org/repo>' to add one")
		return nil
	}

	var errors []string
	for name, repo := range ws.Repos {
		repoDir := filepath.Join(wsPath, repo.Path)
		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			fmt.Printf("  [skip] %s — directory missing\n", name)
			continue
		}

		fmt.Printf("Syncing %s...\n", name)
		if err := git.Pull(repoDir); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
			fmt.Printf("  [fail] %s\n", name)
		} else {
			fmt.Printf("  [ok]   %s\n", name)
		}
	}

	if len(errors) > 0 {
		fmt.Printf("\n%d repo(s) failed to sync:\n", len(errors))
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
		return fmt.Errorf("sync completed with errors")
	}

	fmt.Printf("\nAll %d repos synced\n", len(ws.Repos))
	return nil
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
