package spkconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ConsumesEntry is one model dependency (consumer repo declares "I consume this model").
type ConsumesEntry struct {
	Model   string `json:"model"`
	Package string `json:"package"`
	Codegen string `json:"codegen"`
}

// Config is the per-repo spk.config.json (consumer-centric: repo lists what it consumes).
type Config struct {
	Consumes []ConsumesEntry `json:"consumes"`
}

const ConfigFilename = "spk.config.json"

// Load reads spk.config.json from repoDir. Missing file or empty consumes returns nil, nil.
func Load(repoDir string) (*Config, error) {
	path := filepath.Join(repoDir, ConfigFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
