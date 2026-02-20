package cmd

import (
	"fmt"

	"github.com/Spark-Rewards/homebrew-spk/internal/aws"
	"github.com/Spark-Rewards/homebrew-spk/internal/workspace"
	"github.com/spf13/cobra"
)

var loginProfile string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to AWS SSO using the workspace's profile",
	Long: `Wraps 'aws sso login' using the AWS profile configured in the workspace.
Falls back to the --profile flag if provided.

Examples:
  spk login                  # uses workspace aws_profile
  spk login --profile prod   # override with specific profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := aws.CheckCLI(); err != nil {
			return err
		}

		profile := loginProfile

		// Try to get profile from workspace if not explicitly set
		if profile == "" {
			wsPath, err := workspace.Find()
			if err == nil {
				ws, err := workspace.Load(wsPath)
				if err == nil && ws.AWSProfile != "" {
					profile = ws.AWSProfile
					fmt.Printf("Using workspace AWS profile: %s\n", profile)
				}
			}
		}

		if profile == "" {
			fmt.Println("No AWS profile configured — running 'aws sso login' with default profile")
		}

		fmt.Println("Logging in to AWS SSO...")
		if err := aws.SSOLogin(profile); err != nil {
			return fmt.Errorf("AWS SSO login failed: %w", err)
		}

		fmt.Println("Login successful — verifying identity...")
		return aws.GetCallerIdentity(profile)
	},
}

func init() {
	loginCmd.Flags().StringVar(&loginProfile, "profile", "", "AWS profile to use (overrides workspace setting)")
	rootCmd.AddCommand(loginCmd)
}
