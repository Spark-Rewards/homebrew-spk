package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Spark-Rewards/homebrew-spk/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	testAll   bool
	testWatch bool
)

var knownTestCommands = map[string]string{
	"AppAPI":      "npm test",
	"BusinessAPI": "npm test",
}

var testCmd = &cobra.Command{
	Use:   "test [repo-name]",
	Short: "Run tests for a repo",
	Long: `Runs the test command for a repo. Auto-detects the appropriate test
command based on repo type, or uses test_command from workspace.json.

Examples:
  spk test AppAPI              # run tests for AppAPI
  spk test AppAPI --watch      # run tests in watch mode
  spk test --all               # run tests for all repos`,
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

		if testAll {
			return testAllRepos(wsPath, ws)
		}

		if len(args) == 0 {
			return fmt.Errorf("specify a repo name or use --all")
		}

		return testRepo(wsPath, ws, args[0])
	},
}

func getTestCommand(name string, repo workspace.RepoDef, repoDir string) string {
	if repo.TestCommand != "" {
		return repo.TestCommand
	}

	if cmd, ok := knownTestCommands[name]; ok {
		if testWatch {
			return cmd + ":watch"
		}
		return cmd
	}

	if fileExists(filepath.Join(repoDir, "package.json")) {
		if testWatch {
			return "npm run test:watch"
		}
		return "npm test"
	}

	if fileExists(filepath.Join(repoDir, "build.gradle")) || fileExists(filepath.Join(repoDir, "build.gradle.kts")) {
		return "./gradlew test"
	}

	if fileExists(filepath.Join(repoDir, "go.mod")) {
		return "go test ./..."
	}

	return ""
}

func testRepo(wsPath string, ws *workspace.Workspace, name string) error {
	repo, ok := ws.Repos[name]
	if !ok {
		return fmt.Errorf("repo '%s' not found in workspace", name)
	}

	repoDir := filepath.Join(wsPath, repo.Path)
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("repo directory %s does not exist", repoDir)
	}

	testCmd := getTestCommand(name, repo, repoDir)
	if testCmd == "" {
		fmt.Printf("No test command for '%s' — skipping\n", name)
		return nil
	}

	fmt.Printf("Testing %s: %s\n", name, testCmd)
	return runShell(repoDir, testCmd)
}

func testAllRepos(wsPath string, ws *workspace.Workspace) error {
	if len(ws.Repos) == 0 {
		fmt.Println("No repos in workspace")
		return nil
	}

	var tested, skipped int
	var failures []string

	for name, repo := range ws.Repos {
		repoDir := filepath.Join(wsPath, repo.Path)
		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			fmt.Printf("[skip] %s (not cloned)\n", name)
			skipped++
			continue
		}

		testCmd := getTestCommand(name, repo, repoDir)
		if testCmd == "" {
			fmt.Printf("[skip] %s (no test command)\n", name)
			skipped++
			continue
		}

		fmt.Printf("\n--- Testing %s ---\n", name)
		if err := runShell(repoDir, testCmd); err != nil {
			failures = append(failures, name)
			fmt.Printf("[fail] %s\n", name)
		} else {
			fmt.Printf("[ok]   %s\n", name)
			tested++
		}
	}

	fmt.Printf("\n%d tested, %d skipped", tested, skipped)
	if len(failures) > 0 {
		fmt.Printf(", %d failed: %v", len(failures), failures)
	}
	fmt.Println()

	if len(failures) > 0 {
		return fmt.Errorf("tests failed in %d repo(s)", len(failures))
	}
	return nil
}

func init() {
	testCmd.Flags().BoolVar(&testAll, "all", false, "Test all repos")
	testCmd.Flags().BoolVar(&testWatch, "watch", false, "Run tests in watch mode")
	rootCmd.AddCommand(testCmd)
}
