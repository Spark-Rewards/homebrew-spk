package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Clone clones a repository into the target directory
func Clone(remote, targetDir string) error {
	cmd := exec.Command("git", "clone", remote, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Pull runs git pull in the given directory
func Pull(repoDir string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Status runs git status in the given directory and returns the output
func Status(repoDir string) (string, error) {
	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// CurrentBranch returns the current branch name
func CurrentBranch(repoDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// IsRepo checks if the given directory is a git repository
func IsRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// BuildRemoteURL constructs a git SSH URL from org/repo
func BuildRemoteURL(orgRepo string) string {
	if strings.HasPrefix(orgRepo, "git@") || strings.HasPrefix(orgRepo, "https://") {
		return orgRepo
	}
	return fmt.Sprintf("git@github.com:%s.git", orgRepo)
}

// RepoNameFromRemote extracts the repo name from a remote URL or org/repo string
func RepoNameFromRemote(remote string) string {
	// Handle org/repo format
	if !strings.Contains(remote, ":") && strings.Contains(remote, "/") {
		parts := strings.Split(remote, "/")
		return parts[len(parts)-1]
	}

	// Handle git@github.com:org/repo.git
	base := filepath.Base(remote)
	return strings.TrimSuffix(base, ".git")
}

// Fetch runs git fetch for the specified remote
func Fetch(repoDir, remote string) error {
	if remote == "" {
		remote = "origin"
	}
	cmd := exec.Command("git", "fetch", remote)
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Rebase runs git rebase on the specified upstream branch
func Rebase(repoDir, upstream string) error {
	cmd := exec.Command("git", "rebase", upstream)
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RebaseAbort aborts an in-progress rebase
func RebaseAbort(repoDir string) error {
	cmd := exec.Command("git", "rebase", "--abort")
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Stash stashes uncommitted changes
func Stash(repoDir string) error {
	cmd := exec.Command("git", "stash", "push", "-m", "spk-sync-autostash")
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// StashPop pops the most recent stash
func StashPop(repoDir string) error {
	cmd := exec.Command("git", "stash", "pop")
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// HasStash checks if there are any stashed changes
func HasStash(repoDir string) bool {
	cmd := exec.Command("git", "stash", "list")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

// IsDirty checks if the working directory has uncommitted changes
func IsDirty(repoDir string) bool {
	status, err := Status(repoDir)
	if err != nil {
		return false
	}
	return status != ""
}

// GetDefaultBranch attempts to determine the default branch (main or master)
func GetDefaultBranch(repoDir string) string {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err == nil {
		ref := strings.TrimSpace(string(out))
		parts := strings.Split(ref, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	for _, branch := range []string{"main", "master"} {
		cmd := exec.Command("git", "rev-parse", "--verify", "origin/"+branch)
		cmd.Dir = repoDir
		if err := cmd.Run(); err == nil {
			return branch
		}
	}

	return "main"
}
