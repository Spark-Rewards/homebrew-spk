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
	Remote        string   `json:"remote"`
	Path          string   `json:"path"`
	BuildCommand  string   `json:"build_command,omitempty"`
	TestCommand   string   `json:"test_command,omitempty"`
	Dependencies  []string `json:"dependencies,omitempty"`
	DefaultBranch string   `json:"default_branch,omitempty"`
	ModelFor      string   `json:"model_for,omitempty"`
}

type Workspace struct {
	Name          string             `json:"name"`
	CreatedAt     string             `json:"created_at"`
	AWSProfile    string             `json:"aws_profile,omitempty"`
	AWSRegion     string             `json:"aws_region,omitempty"`
	Repos         map[string]RepoDef `json:"repos"`
	Env           map[string]string  `json:"env,omitempty"`
	DefaultBranch string             `json:"default_branch,omitempty"`
	SSMEnvPath    string             `json:"ssm_env_path,omitempty"`
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

	return "", fmt.Errorf("not inside a spark-cli workspace (no .spk/workspace.json found)")
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

// VSCodeWorkspacePath returns the path to the .code-workspace file
func VSCodeWorkspacePath(workspacePath string) string {
	ws, err := Load(workspacePath)
	if err != nil {
		return filepath.Join(workspacePath, "workspace.code-workspace")
	}
	return filepath.Join(workspacePath, ws.Name+".code-workspace")
}

// GenerateVSCodeWorkspace creates/updates the .code-workspace file
func GenerateVSCodeWorkspace(workspacePath string) error {
	ws, err := Load(workspacePath)
	if err != nil {
		return err
	}

	type folder struct {
		Path string `json:"path"`
	}
	type vscodeWorkspace struct {
		Folders []folder `json:"folders"`
	}

	var folders []folder
	for _, repo := range ws.Repos {
		folders = append(folders, folder{Path: repo.Path})
	}

	vscWs := vscodeWorkspace{Folders: folders}

	data, err := json.MarshalIndent(vscWs, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal VS Code workspace: %w", err)
	}

	wsFile := VSCodeWorkspacePath(workspacePath)
	return os.WriteFile(wsFile, data, 0644)
}

// GlobalEnvPath returns the path to the workspace's global .env file
func GlobalEnvPath(workspacePath string) string {
	return filepath.Join(workspacePath, ".env")
}

// WriteGlobalEnv writes environment variables to the workspace's global .env file
func WriteGlobalEnv(workspacePath string, vars map[string]string) error {
	envPath := GlobalEnvPath(workspacePath)

	existing, _ := ReadGlobalEnv(workspacePath)
	if existing == nil {
		existing = make(map[string]string)
	}

	for k, v := range vars {
		existing[k] = v
	}

	var lines []string
	for k, v := range existing {
		lines = append(lines, fmt.Sprintf("%s=%s", k, v))
	}

	content := ""
	for _, line := range lines {
		content += line + "\n"
	}

	return os.WriteFile(envPath, []byte(content), 0644)
}

// ReadGlobalEnv reads the workspace's global .env file into a map
func ReadGlobalEnv(workspacePath string) (map[string]string, error) {
	envPath := GlobalEnvPath(workspacePath)

	data, err := os.ReadFile(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("failed to read .env file: %w", err)
	}

	result := make(map[string]string)
	lines := splitLines(string(data))

	for _, line := range lines {
		line = trimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}

		idx := indexByte(line, '=')
		if idx == -1 {
			continue
		}

		key := line[:idx]
		value := line[idx+1:]
		result[key] = value
	}

	return result, nil
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
