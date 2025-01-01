// internal/config/types.go
package config

// Config represents the backup configuration
type Config struct {
	SourceDirectory string   `yaml:"source_directory"`
	TargetDirectory string   `yaml:"target_directory"`
	FoldersToBackup []string `yaml:"folders_to_backup"`
	Concurrency     int      `yaml:"concurrency"`
	BufferSize      int      `yaml:"buffer_size"`
	RetryAttempts   int      `yaml:"retry_attempts"`
	RetryDelay      string   `yaml:"retry_delay"`
	ExcludePatterns []string `yaml:"exclude_patterns"`
	LogLevel        string   `yaml:"log_level"`
}

type ConfigOptions struct {
	Quiet    bool   `yaml:"quiet"`
	Verbose  bool   `yaml:"verbose"`
	LogLevel string `yaml:"log_level"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Concurrency:   2,
		BufferSize:    32 * 1024, // 32KB
		RetryAttempts: 3,
		RetryDelay:    "1s",
		LogLevel:      "info",
		ExcludePatterns: []string{
			"*.tmp",
			".DS_Store",
			"Thumbs.db",
		},
	}
}
