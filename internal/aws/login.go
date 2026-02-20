package aws

import (
	"fmt"
	"os"
	"os/exec"
)

// SSOLogin runs `aws sso login` with the given profile
func SSOLogin(profile string) error {
	args := []string{"sso", "login"}
	if profile != "" {
		args = append(args, "--profile", profile)
	}

	cmd := exec.Command("aws", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// CheckCLI verifies that the AWS CLI is installed
func CheckCLI() error {
	_, err := exec.LookPath("aws")
	if err != nil {
		return fmt.Errorf("AWS CLI not found — install it with: brew install awscli")
	}
	return nil
}

// GetCallerIdentity runs `aws sts get-caller-identity` to verify credentials
func GetCallerIdentity(profile string) error {
	args := []string{"sts", "get-caller-identity"}
	if profile != "" {
		args = append(args, "--profile", profile)
	}

	cmd := exec.Command("aws", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
