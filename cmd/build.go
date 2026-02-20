package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Spark-Rewards/homebrew-spk/internal/npm"
	"github.com/Spark-Rewards/homebrew-spk/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	buildAll      bool
	buildNoLink   bool
	buildPublished bool
)

var knownBuildCommands = map[string]string{
	"AppModel":      "npm run build:all",
	"BusinessModel": "npm run build:all",
	"AppAPI":        "npm run build",
	"BusinessAPI":   "npm run build",
}

type depMapping struct {
	api string
	pkg string
}

var modelToAPI = map[string]depMapping{
	"AppModel":      {api: "AppAPI", pkg: "@spark-rewards/sra-sdk"},
	"BusinessModel": {api: "BusinessAPI", pkg: "@spark-rewards/srw-sdk"},
}

var apiToModel = map[string]string{
	"AppAPI":      "AppModel",
	"BusinessAPI": "BusinessModel",
}

var buildCmd = &cobra.Command{
	Use:   "build [repo-name]",
	Short: "Build a repo with automatic local dependency linking",
	Long: `Builds a repo and automatically links locally-built dependencies.

Like Amazon's Brazil Build, spk automatically detects when a dependency
(like a Smithy model) is built locally and links it to consuming packages
(like APIs) instead of using published versions.

Dependency chain:
  AppModel      -> AppAPI      (@spark-rewards/sra-sdk)
  BusinessModel -> BusinessAPI (@spark-rewards/srw-sdk)

When you build an API, spk checks if its model is built locally:
  - If YES: links the local build via npm link (live development)
  - If NO:  uses the published package from npm registry

Examples:
  spk build AppModel           # build model, auto-link to AppAPI if present
  spk build AppAPI             # build API, auto-link local AppModel if built
  spk build --all              # build all in dependency order with linking
  spk build AppAPI --published # force use of published packages (no linking)`,
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

func getBuildCommand(name string, repo workspace.RepoDef, repoDir string) string {
	if repo.BuildCommand != "" {
		return repo.BuildCommand
	}

	if cmd, ok := knownBuildCommands[name]; ok {
		return cmd
	}

	if fileExists(filepath.Join(repoDir, "package.json")) {
		return "npm run build"
	}
	if fileExists(filepath.Join(repoDir, "build.gradle")) || fileExists(filepath.Join(repoDir, "build.gradle.kts")) {
		return "./gradlew build"
	}
	if fileExists(filepath.Join(repoDir, "Makefile")) {
		return "make"
	}
	if fileExists(filepath.Join(repoDir, "go.mod")) {
		return "go build ./..."
	}

	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func buildRepo(wsPath string, ws *workspace.Workspace, name string) error {
	repo, ok := ws.Repos[name]
	if !ok {
		return fmt.Errorf("repo '%s' not found in workspace", name)
	}

	repoDir := filepath.Join(wsPath, repo.Path)
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("repo directory %s does not exist", repoDir)
	}

	fmt.Printf("=== Building %s ===\n", name)

	if !buildNoLink && !buildPublished {
		if err := autoLinkDependencies(wsPath, ws, name); err != nil {
			fmt.Printf("Warning: dependency linking issue: %v\n", err)
		}
	}

	buildCmd := getBuildCommand(name, repo, repoDir)
	if buildCmd == "" {
		fmt.Printf("No build command for '%s' — skipping\n", name)
		return nil
	}

	fmt.Printf("Running: %s\n", buildCmd)
	if err := runShell(repoDir, buildCmd); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	if !buildNoLink && !buildPublished {
		if err := autoLinkToConsumers(wsPath, ws, name); err != nil {
			fmt.Printf("Note: %v\n", err)
		}
	}

	fmt.Printf("[ok] %s built successfully\n", name)
	return nil
}

func autoLinkDependencies(wsPath string, ws *workspace.Workspace, name string) error {
	modelName, isAPI := apiToModel[name]
	if !isAPI {
		return nil
	}

	modelRepo, exists := ws.Repos[modelName]
	if !exists {
		return nil
	}

	modelDir := filepath.Join(wsPath, modelRepo.Path)
	apiDir := filepath.Join(wsPath, ws.Repos[name].Path)
	mapping := modelToAPI[modelName]

	if !npm.IsBuilt(modelDir) {
		fmt.Printf("Using published %s (local not built)\n", mapping.pkg)
		return nil
	}

	if npm.IsLinked(apiDir, mapping.pkg) {
		fmt.Printf("Using local %s (already linked)\n", modelName)
		return nil
	}

	fmt.Printf("Linking local %s -> %s...\n", modelName, name)
	buildDir := npm.BuildOutputDir(modelDir)

	if err := npm.Link(buildDir); err != nil {
		return fmt.Errorf("npm link in %s failed: %w", modelName, err)
	}

	if err := npm.LinkPackage(apiDir, mapping.pkg); err != nil {
		return fmt.Errorf("npm link %s failed: %w", mapping.pkg, err)
	}

	fmt.Printf("Linked: %s now uses local %s\n", name, modelName)
	return nil
}

func autoLinkToConsumers(wsPath string, ws *workspace.Workspace, name string) error {
	mapping, isModel := modelToAPI[name]
	if !isModel {
		return nil
	}

	apiRepo, exists := ws.Repos[mapping.api]
	if !exists {
		return nil
	}

	apiDir := filepath.Join(wsPath, apiRepo.Path)
	if _, err := os.Stat(apiDir); os.IsNotExist(err) {
		return nil
	}

	modelDir := filepath.Join(wsPath, ws.Repos[name].Path)
	buildDir := npm.BuildOutputDir(modelDir)

	if !npm.IsBuilt(modelDir) {
		return nil
	}

	if npm.IsLinked(apiDir, mapping.pkg) {
		return nil
	}

	fmt.Printf("Auto-linking to consumer %s...\n", mapping.api)

	if err := npm.Link(buildDir); err != nil {
		return fmt.Errorf("npm link failed: %w", err)
	}

	if err := npm.LinkPackage(apiDir, mapping.pkg); err != nil {
		return fmt.Errorf("npm link %s in %s failed: %w", mapping.pkg, mapping.api, err)
	}

	fmt.Printf("Linked: %s now uses local %s\n", mapping.api, name)
	return nil
}

func buildAllRepos(wsPath string, ws *workspace.Workspace) error {
	order := getSmartBuildOrder(ws)

	fmt.Printf("Build order: %v\n", order)
	fmt.Printf("Local linking: %v\n\n", !buildNoLink && !buildPublished)

	for _, name := range order {
		repo, exists := ws.Repos[name]
		if !exists {
			continue
		}

		repoDir := filepath.Join(wsPath, repo.Path)
		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			fmt.Printf("[skip] %s (not cloned)\n\n", name)
			continue
		}

		if err := buildRepo(wsPath, ws, name); err != nil {
			return fmt.Errorf("build failed at '%s': %w", name, err)
		}
		fmt.Println()
	}

	fmt.Println("All builds completed")
	return nil
}

func getSmartBuildOrder(ws *workspace.Workspace) []string {
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)

	for name := range ws.Repos {
		inDegree[name] = 0
	}

	for name := range ws.Repos {
		if modelName, isAPI := apiToModel[name]; isAPI {
			if _, modelExists := ws.Repos[modelName]; modelExists {
				dependents[modelName] = append(dependents[modelName], name)
				inDegree[name]++
			}
		}
	}

	for name, repo := range ws.Repos {
		for _, dep := range repo.Dependencies {
			if _, exists := ws.Repos[dep]; exists {
				alreadyAdded := false
				for _, d := range dependents[dep] {
					if d == name {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					dependents[dep] = append(dependents[dep], name)
					inDegree[name]++
				}
			}
		}
	}

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

	for name := range ws.Repos {
		found := false
		for _, o := range order {
			if o == name {
				found = true
				break
			}
		}
		if !found {
			order = append(order, name)
		}
	}

	return order
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
	buildCmd.Flags().BoolVar(&buildNoLink, "no-link", false, "Disable automatic local dependency linking")
	buildCmd.Flags().BoolVar(&buildPublished, "published", false, "Force use of published packages (no local linking)")
	rootCmd.AddCommand(buildCmd)
}
