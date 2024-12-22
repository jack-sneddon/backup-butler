package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
	"gopkg.in/yaml.v3"
)

type FileConfigLoader struct {
	// Default values
	defaultConcurrency   int
	defaultBufferSize    int
	defaultRetryAttempts int
	defaultRetryDelay    time.Duration
	defaultChecksumAlg   string
}

func NewFileConfigLoader() *FileConfigLoader {
	return &FileConfigLoader{
		defaultConcurrency:   4,
		defaultBufferSize:    32 * 1024, // 32KB
		defaultRetryAttempts: 3,
		defaultRetryDelay:    time.Second,
		defaultChecksumAlg:   "sha256",
	}
}

func (l *FileConfigLoader) Load(path string) (*backup.BackupConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", path, err)
	}

	// Define temporary struct for parsing
	type tempConfig struct {
		SourceDirectory    string   `json:"source_directory" yaml:"source_directory"`
		TargetDirectory    string   `json:"target_directory" yaml:"target_directory"`
		FoldersToBackup    []string `json:"folders_to_backup" yaml:"folders_to_backup"`
		DeepDuplicateCheck bool     `json:"deep_duplicate_check" yaml:"deep_duplicate_check"`
		Concurrency        int      `json:"concurrency" yaml:"concurrency"`
		BufferSize         int      `json:"buffer_size" yaml:"buffer_size"`
		RetryAttempts      int      `json:"retry_attempts" yaml:"retry_attempts"`
		RetryDelay         string   `json:"retry_delay" yaml:"retry_delay"` // String for parsing
		ExcludePatterns    []string `json:"exclude_patterns" yaml:"exclude_patterns"`
		ChecksumAlgorithm  string   `json:"checksum_algorithm" yaml:"checksum_algorithm"`
		LogLevel           string   `json:"log_level" yaml:"log_level"`
	}

	// Parse into temporary struct
	var temp tempConfig

	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &temp); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config file: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &temp); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config file: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config format '%s', must be .json, .yaml, or .yml", ext)
	}

	// Create final config with defaults
	config := &backup.BackupConfig{
		SourceDirectory:    temp.SourceDirectory,
		TargetDirectory:    temp.TargetDirectory,
		FoldersToBackup:    temp.FoldersToBackup,
		DeepDuplicateCheck: temp.DeepDuplicateCheck,
		Concurrency:        temp.Concurrency,
		BufferSize:         temp.BufferSize,
		RetryAttempts:      temp.RetryAttempts,
		ChecksumAlgorithm:  temp.ChecksumAlgorithm,
		ExcludePatterns:    temp.ExcludePatterns,
		LogLevel:           temp.LogLevel,
		RetryDelay:         l.defaultRetryDelay, // Default value
		Options:            &backup.ConfigOptions{},
	}

	// Parse retry delay if provided
	if temp.RetryDelay != "" {
		duration, err := time.ParseDuration(temp.RetryDelay)
		if err != nil {
			return nil, fmt.Errorf("invalid retry_delay format '%s': must be a valid duration (e.g., '1s', '500ms')", temp.RetryDelay)
		}
		config.RetryDelay = duration
	}

	// Set up Options
	config.Options.LogLevel = config.LogLevel

	// Validate the configuration
	if err := l.Validate(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (l *FileConfigLoader) Validate(config *backup.BackupConfig) error {
	var errors []string

	// Required fields validation
	if config.SourceDirectory == "" {
		errors = append(errors, "source_directory is required")
	}
	if config.TargetDirectory == "" {
		errors = append(errors, "target_directory is required")
	}
	if len(config.FoldersToBackup) == 0 {
		errors = append(errors, "folders_to_backup must contain at least one folder")
	}

	// Source directory existence
	if _, err := os.Stat(config.SourceDirectory); err != nil {
		if os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("source_directory '%s' does not exist", config.SourceDirectory))
		} else {
			errors = append(errors, fmt.Sprintf("error accessing source_directory '%s': %v", config.SourceDirectory, err))
		}
	}

	// Numeric value validation
	if config.Concurrency < 1 {
		errors = append(errors, fmt.Sprintf("concurrency must be at least 1 (got %d)", config.Concurrency))
	} else if config.Concurrency > 32 {
		errors = append(errors, fmt.Sprintf("concurrency must not exceed 32 (got %d)", config.Concurrency))
	}

	if config.BufferSize < 1024 {
		errors = append(errors, fmt.Sprintf("buffer_size must be at least 1024 bytes (got %d)", config.BufferSize))
	} else if config.BufferSize > 10*1024*1024 {
		errors = append(errors, fmt.Sprintf("buffer_size must not exceed 10MB (got %d)", config.BufferSize))
	}

	if config.RetryAttempts < 0 {
		errors = append(errors, fmt.Sprintf("retry_attempts cannot be negative (got %d)", config.RetryAttempts))
	}

	// Checksum algorithm validation
	validAlgorithms := map[string]bool{"sha256": true, "sha512": true, "md5": true}
	if !validAlgorithms[config.ChecksumAlgorithm] {
		errors = append(errors, fmt.Sprintf("checksum_algorithm must be one of: sha256, sha512, md5 (got '%s')", config.ChecksumAlgorithm))
	}

	// Log level validation
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[config.LogLevel] {
		errors = append(errors, fmt.Sprintf("log_level must be one of: debug, info, warn, error (got '%s')", config.LogLevel))
	}

	// Exclude patterns validation
	for _, pattern := range config.ExcludePatterns {
		if _, err := filepath.Match(pattern, "test"); err != nil {
			errors = append(errors, fmt.Sprintf("invalid exclude pattern '%s': %v", pattern, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n- %s", strings.Join(errors, "\n- "))
	}

	return nil
}
