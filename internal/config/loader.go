// internal/config/loader.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads and validates the configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func validateConfig(config *Config) error {
	if config.SourceDirectory == "" {
		return fmt.Errorf("source_directory is required")
	}

	if config.TargetDirectory == "" {
		return fmt.Errorf("target_directory is required")
	}

	if len(config.FoldersToBackup) == 0 {
		return fmt.Errorf("folders_to_backup must contain at least one folder")
	}

	// Validate source directory exists
	if _, err := os.Stat(config.SourceDirectory); err != nil {
		return fmt.Errorf("source directory error: %w", err)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(config.TargetDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Validate numeric settings
	if config.Concurrency < 1 {
		config.Concurrency = 1
	}
	if config.Concurrency > 8 {
		config.Concurrency = 8
	}

	if config.BufferSize < 4096 {
		config.BufferSize = 4096
	}
	if config.BufferSize > 10*1024*1024 {
		config.BufferSize = 10 * 1024 * 1024
	}

	// Validate retry delay
	if _, err := time.ParseDuration(config.RetryDelay); err != nil {
		return fmt.Errorf("invalid retry_delay format: %w", err)
	}

	// Validate exclude patterns
	for _, pattern := range config.ExcludePatterns {
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return fmt.Errorf("invalid exclude pattern '%s': %w", pattern, err)
		}
	}

	return nil
}

func parseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "quiet":
		return LogQuiet
	case "verbose":
		return LogVerbose
	case "debug":
		return LogDebug
	default:
		return LogNormal
	}
}
