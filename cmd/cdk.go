package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Spark-Rewards/homebrew-spark-cli/internal/workspace"
	"github.com/spf13/cobra"
)

const cdkConfigFile = "cdk.json"

var cdkCmd = &cobra.Command{
	Use:   "cdk [cdk-args...]",
	Short: "Run AWS CDK CLI in the workspace CDK repo (e.g. list, deploy, diff | -h)",
	Long: `Runs the AWS CDK CLI in the workspace context. Resolves the CDK app directory
from the current repo (if it contains cdk.json) or from CorePipeline (or any
workspace repo that contains cdk.json). Passes all arguments through to cdk.

Examples:
  spark-cli cdk list
  spark-cli cdk deploy PipelineStack/beta/SomeStack
  spark-cli cdk diff
  spark-cli cdk synth`,
	Args:            cobra.ArbitraryArgs,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		wsPath, err := workspace.Find()
		if err != nil {
			return err
		}

		ws, err := workspace.Load(wsPath)
		if err != nil {
			return err
		}

		cdkDir, err := findCDKRepoDir(wsPath, ws)
		if err != nil {
			return err
		}

		cdkPath, err := exec.LookPath("cdk")
		if err != nil {
			return fmt.Errorf("cdk not found in PATH — install with: npm install -g aws-cdk")
		}

		c := exec.Command(cdkPath, args...)
		c.Dir = cdkDir
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Env = os.Environ()

		if err := c.Run(); err != nil {
			if exit, ok := err.(*exec.ExitError); ok {
				os.Exit(exit.ExitCode())
			}
			return err
		}
		return nil
	},
}

// findCDKRepoDir returns the repo directory that contains cdk.json.
// Prefers the repo containing the current working dir; otherwise the first workspace repo with cdk.json (e.g. CorePipeline).
func findCDKRepoDir(wsPath string, ws *workspace.Workspace) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// If cwd is inside a repo that has cdk.json, use it.
	for _, repo := range ws.Repos {
		repoDir := filepath.Join(wsPath, repo.Path)
		absRepo, _ := filepath.Abs(repoDir)
		if cwd == absRepo || isSubdir(absRepo, cwd) {
			if hasCDK(repoDir) {
				return repoDir, nil
			}
			break
		}
	}

	// Else use first workspace repo that has cdk.json (e.g. CorePipeline).
	for _, repo := range ws.Repos {
		repoDir := filepath.Join(wsPath, repo.Path)
		if hasCDK(repoDir) {
			return repoDir, nil
		}
	}

	return "", fmt.Errorf("no CDK app (cdk.json) found in workspace — run from CorePipeline or add cdk.json to a repo")
}

func hasCDK(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, cdkConfigFile))
	return err == nil
}

func init() {
	rootCmd.AddCommand(cdkCmd)
}
