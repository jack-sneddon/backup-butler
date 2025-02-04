// internal/config/loader.go
package config

import (
	"fmt"
	"os"

	"github.com/jack-sneddon/backup-butler/internal/scan"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	cfg.SetDefaults()

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize empty critical paths if not set
	if cfg.Validation != nil && cfg.Validation.CriticalPaths == nil {
		cfg.Validation.CriticalPaths = make([]scan.CriticalPath, 0)
	}

	return &cfg, nil
}
