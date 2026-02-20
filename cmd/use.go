package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Spark-Rewards/homebrew-spk/internal/config"
	"github.com/Spark-Rewards/homebrew-spk/internal/git"
	"github.com/Spark-Rewards/homebrew-spk/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	useBuildCmd string
	useDeps     []string
)

var useCmd = &cobra.Command{
	Use:   "use <org/repo | repo-url>",
	Short: "Clone a repo into the current workspace",
	Long: `Clones a GitHub repository into the current workspace and registers it
in the workspace manifest.

If only a repo name is provided (no org), the default_github_org from
~/.spk/config.json is used.

Examples:
  spk use my-org/BusinessAPI
  spk use BusinessAPI                              # uses default org
  spk use git@github.com:my-org/BusinessAPI.git    # full URL`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoArg := args[0]

		// Find workspace
		wsPath, err := workspace.Find()
		if err != nil {
			return fmt.Errorf("you must be inside a spk workspace — run 'spk create workspace <path>' first")
		}

		// Resolve the remote URL
		remote := resolveRemote(repoArg)
		repoName := git.RepoNameFromRemote(repoArg)
		targetDir := filepath.Join(wsPath, repoName)

		// Check if already cloned
		if _, err := os.Stat(targetDir); err == nil {
			if git.IsRepo(targetDir) {
				fmt.Printf("Repository '%s' already exists at %s\n", repoName, targetDir)
				// Still register it in manifest if not present
				return registerRepo(wsPath, repoName, remote, targetDir)
			}
			return fmt.Errorf("directory %s exists but is not a git repository", targetDir)
		}

		// Clone
		fmt.Printf("Cloning %s into %s...\n", remote, targetDir)
		if err := git.Clone(remote, targetDir); err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}

		// Register in workspace manifest
		if err := registerRepo(wsPath, repoName, remote, targetDir); err != nil {
			return err
		}

		fmt.Printf("Repository '%s' added to workspace\n", repoName)
		return nil
	},
}

func resolveRemote(arg string) string {
	// If it's already a full URL, use as-is
	if git.BuildRemoteURL(arg) == arg {
		return arg
	}

	// If no slash, try to prepend default org
	if !containsSlash(arg) {
		cfg, err := config.LoadGlobal()
		if err == nil && cfg.DefaultGithubOrg != "" {
			return git.BuildRemoteURL(cfg.DefaultGithubOrg + "/" + arg)
		}
	}

	return git.BuildRemoteURL(arg)
}

func containsSlash(s string) bool {
	for _, c := range s {
		if c == '/' {
			return true
		}
	}
	return false
}

func registerRepo(wsPath, name, remote, targetDir string) error {
	relPath, _ := filepath.Rel(wsPath, targetDir)
	repo := workspace.RepoDef{
		Remote:       remote,
		Path:         relPath,
		BuildCommand: useBuildCmd,
		Dependencies: useDeps,
	}
	return workspace.AddRepo(wsPath, name, repo)
}

func init() {
	useCmd.Flags().StringVar(&useBuildCmd, "build", "", "Build command for this repo (e.g., 'npm run build')")
	useCmd.Flags().StringSliceVar(&useDeps, "deps", nil, "Dependencies (other repo names that must build first)")
	rootCmd.AddCommand(useCmd)
}
