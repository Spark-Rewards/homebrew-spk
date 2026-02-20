package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	SparkDir       = ".spk"
	GlobalFileName = "config.json"
)

type GlobalConfig struct {
	DefaultGithubOrg string   `json:"default_github_org"`
	DefaultAWSProfile string  `json:"default_aws_profile"`
	DefaultAWSRegion  string  `json:"default_aws_region"`
	Workspaces       []string `json:"workspaces"`
}

// GlobalDir returns ~/.spk
func GlobalDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not find home directory: %w", err)
	}
	return filepath.Join(home, SparkDir), nil
}

// GlobalConfigPath returns ~/.spk/config.json
func GlobalConfigPath() (string, error) {
	dir, err := GlobalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, GlobalFileName), nil
}

// EnsureGlobalDir creates ~/.spk if it doesn't exist
func EnsureGlobalDir() error {
	dir, err := GlobalDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

// LoadGlobal reads the global config from ~/.spk/config.json
func LoadGlobal() (*GlobalConfig, error) {
	path, err := GlobalConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &GlobalConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read global config: %w", err)
	}

	var cfg GlobalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse global config: %w", err)
	}
	return &cfg, nil
}

// SaveGlobal writes the global config to ~/.spk/config.json
func SaveGlobal(cfg *GlobalConfig) error {
	if err := EnsureGlobalDir(); err != nil {
		return err
	}

	path, err := GlobalConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal global config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// RegisterWorkspace adds a workspace path to the global config if not already present
func RegisterWorkspace(absPath string) error {
	cfg, err := LoadGlobal()
	if err != nil {
		return err
	}

	for _, ws := range cfg.Workspaces {
		if ws == absPath {
			return nil // already registered
		}
	}

	cfg.Workspaces = append(cfg.Workspaces, absPath)
	return SaveGlobal(cfg)
}

// SetDefaults updates the global config with provided defaults
func SetDefaults(org, awsProfile, awsRegion string) error {
	cfg, err := LoadGlobal()
	if err != nil {
		return err
	}

	if org != "" {
		cfg.DefaultGithubOrg = org
	}
	if awsProfile != "" {
		cfg.DefaultAWSProfile = awsProfile
	}
	if awsRegion != "" {
		cfg.DefaultAWSRegion = awsRegion
	}

	return SaveGlobal(cfg)
}
