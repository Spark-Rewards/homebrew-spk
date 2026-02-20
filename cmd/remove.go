package cmd

import (
	"fmt"

	"github.com/Spark-Rewards/homebrew-spk/internal/workspace"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <repo-name>",
	Short: "Remove a repo from the workspace manifest",
	Long: `Unregisters a repo from workspace.json. This does NOT delete the
directory — it only removes the entry from the manifest.

Example:
  spk remove BusinessAPI`,
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		wsPath, err := workspace.Find()
		if err != nil {
			return err
		}

		ws, err := workspace.Load(wsPath)
		if err != nil {
			return err
		}

		if _, ok := ws.Repos[name]; !ok {
			return fmt.Errorf("repo '%s' not found in workspace", name)
		}

		if err := workspace.RemoveRepo(wsPath, name); err != nil {
			return err
		}

		fmt.Printf("Removed '%s' from workspace manifest\n", name)
		fmt.Println("(directory was not deleted — remove it manually if needed)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
