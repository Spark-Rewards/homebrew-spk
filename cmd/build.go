package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Spark-Rewards/homebrew-spk/internal/workspace"
	"github.com/spf13/cobra"
)

var buildAll bool

var buildCmd = &cobra.Command{
	Use:   "build [repo-name]",
	Short: "Run the build command for a repo or all repos",
	Long: `Executes the build_command defined in workspace.json for a specific repo,
or all repos in dependency order when --all is used.

Examples:
  spk build BusinessModel             # build one repo
  spk build --all                     # build all in dependency order`,
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

		if buildAll {
			return buildAllRepos(wsPath, ws)
		}

		if len(args) == 0 {
			return fmt.Errorf("specify a repo name or use --all")
		}

		return buildRepo(wsPath, ws, args[0])
	},
}

func buildRepo(wsPath string, ws *workspace.Workspace, name string) error {
	repo, ok := ws.Repos[name]
	if !ok {
		return fmt.Errorf("repo '%s' not found in workspace", name)
	}

	if repo.BuildCommand == "" {
		fmt.Printf("No build_command configured for '%s' — skipping\n", name)
		return nil
	}

	repoDir := filepath.Join(wsPath, repo.Path)
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("repo directory %s does not exist", repoDir)
	}

	fmt.Printf("Building %s: %s\n", name, repo.BuildCommand)
	return runShell(repoDir, repo.BuildCommand)
}

func buildAllRepos(wsPath string, ws *workspace.Workspace) error {
	order, err := topoSort(ws.Repos)
	if err != nil {
		return err
	}

	fmt.Printf("Build order: %v\n\n", order)

	for _, name := range order {
		if err := buildRepo(wsPath, ws, name); err != nil {
			return fmt.Errorf("build failed at '%s': %w", name, err)
		}
		fmt.Println()
	}

	fmt.Println("All builds completed")
	return nil
}

// topoSort performs a topological sort on repos by their dependencies
func topoSort(repos map[string]workspace.RepoDef) ([]string, error) {
	// Build adjacency: repo -> repos that depend on it
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)

	for name := range repos {
		inDegree[name] = 0
	}

	for name, repo := range repos {
		for _, dep := range repo.Dependencies {
			if _, exists := repos[dep]; !exists {
				return nil, fmt.Errorf("repo '%s' depends on '%s' which is not in the workspace", name, dep)
			}
			dependents[dep] = append(dependents[dep], name)
			inDegree[name]++
		}
	}

	// Kahn's algorithm
	var queue []string
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	var order []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)

		for _, dep := range dependents[current] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if len(order) != len(repos) {
		return nil, fmt.Errorf("circular dependency detected in workspace repos")
	}

	return order, nil
}

func runShell(dir, command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func init() {
	buildCmd.Flags().BoolVar(&buildAll, "all", false, "Build all repos in dependency order")
	rootCmd.AddCommand(buildCmd)
}
