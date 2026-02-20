package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Spark-Rewards/homebrew-spk/internal/config"
)

const ManifestFile = "workspace.json"

type RepoDef struct {
	Remote       string   `json:"remote"`
	Path         string   `json:"path"`
	BuildCommand string   `json:"build_command,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

type Workspace struct {
	Name       string             `json:"name"`
	CreatedAt  string             `json:"created_at"`
	AWSProfile string             `json:"aws_profile,omitempty"`
	AWSRegion  string             `json:"aws_region,omitempty"`
	Repos      map[string]RepoDef `json:"repos"`
	Env        map[string]string  `json:"env,omitempty"`
}

// SparkDir returns the .spark directory path within a workspace
func SparkDir(workspacePath string) string {
	return filepath.Join(workspacePath, config.SparkDir)
}

// ManifestPath returns the full path to workspace.json
func ManifestPath(workspacePath string) string {
	return filepath.Join(SparkDir(workspacePath), ManifestFile)
}

// Create initializes a new workspace at the given path
func Create(absPath, name, awsProfile, awsRegion string) (*Workspace, error) {
	sparkDir := SparkDir(absPath)
	if err := os.MkdirAll(sparkDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .spark directory: %w", err)
	}

	ws := &Workspace{
		Name:       name,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		AWSProfile: awsProfile,
		AWSRegion:  awsRegion,
		Repos:      make(map[string]RepoDef),
		Env:        make(map[string]string),
	}

	if awsRegion != "" {
		ws.Env["AWS_REGION"] = awsRegion
	}

	if err := Save(absPath, ws); err != nil {
		return nil, err
	}

	if err := config.RegisterWorkspace(absPath); err != nil {
		return nil, fmt.Errorf("failed to register workspace globally: %w", err)
	}

	return ws, nil
}

// Load reads the workspace manifest from disk
func Load(workspacePath string) (*Workspace, error) {
	path := ManifestPath(workspacePath)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace manifest: %w", err)
	}

	var ws Workspace
	if err := json.Unmarshal(data, &ws); err != nil {
		return nil, fmt.Errorf("failed to parse workspace manifest: %w", err)
	}
	return &ws, nil
}

// Save writes the workspace manifest to disk
func Save(workspacePath string, ws *Workspace) error {
	path := ManifestPath(workspacePath)
	data, err := json.MarshalIndent(ws, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workspace manifest: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// Find walks up from the current directory to find a workspace root
func Find() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		manifest := ManifestPath(dir)
		if _, err := os.Stat(manifest); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("not inside a spk workspace (no .spk/workspace.json found)")
}

// AddRepo registers a repo in the workspace manifest
func AddRepo(workspacePath, name string, repo RepoDef) error {
	ws, err := Load(workspacePath)
	if err != nil {
		return err
	}

	if ws.Repos == nil {
		ws.Repos = make(map[string]RepoDef)
	}
	ws.Repos[name] = repo

	return Save(workspacePath, ws)
}

// RemoveRepo removes a repo from the workspace manifest
func RemoveRepo(workspacePath, name string) error {
	ws, err := Load(workspacePath)
	if err != nil {
		return err
	}

	delete(ws.Repos, name)
	return Save(workspacePath, ws)
}
