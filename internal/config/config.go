package config

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// BackupConfig holds the application configuration
type BackupConfig struct {
	Source          string   `yaml:"source"`
	Target          string   `yaml:"target"`
	IncludeDirs     []string `yaml:"include_dirs"`
	ExcludePatterns []string `yaml:"exclude_patterns"`
	LogDir          string   `yaml:"log_dir"`
}

// LoadConfig loads configuration from YAML file
func LoadConfig(path string) (*BackupConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config BackupConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults if not specified
	if config.LogDir == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			config.LogDir = homeDir + "/backup-logs"
		} else {
			config.LogDir = "./backup-logs"
		}
	} else {
		// Expand $HOME if present in log_dir
		if strings.HasPrefix(config.LogDir, "$HOME") {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				config.LogDir = strings.Replace(config.LogDir, "$HOME", homeDir, 1)
			}
		}
	}

	// Validate required fields
	if config.Source == "" {
		return nil, &ValidationError{"source path is required"}
	}
	if config.Target == "" {
		return nil, &ValidationError{"target path is required"}
	}

	// Ensure paths exist
	if _, err := os.Stat(config.Source); os.IsNotExist(err) {
		return nil, &ValidationError{"source path does not exist"}
	}

	return &config, nil
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return "configuration error: " + e.Message
}