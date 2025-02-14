// internal/config/config.go
package config

import (
	"fmt"
	"os"

	"github.com/jack-sneddon/backup-butler/internal/scan"
)

type Config struct {
	Source  string   `yaml:"source"`
	Target  string   `yaml:"target"`
	Folders []string `yaml:"folders,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`

	Comparison struct {
		Algorithm  string `yaml:"algorithm"`
		Level      string `yaml:"level"`
		BufferSize int    `yaml:"buffer_size"`
	} `yaml:"comparison"`

	Validation *scan.ValidationConfig `yaml:"validation"`

	Storage StorageConfig `yaml:"storage"`

	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`
}

type StorageDeviceConfig struct {
	Type       string `yaml:"type"`        // "hdd", "ssd", "network"
	BufferSize int    `yaml:"buffer_size"` // Optional, use default if 0
	MaxThreads int    `yaml:"max_threads"` // Optional, use default if 0
}

type StorageConfig struct {
	Source struct {
		Type       string `yaml:"type"`        // hdd, ssd, network
		BufferSize int    `yaml:"buffer_size"` // Optional
		MaxThreads int    `yaml:"max_threads"` // Optional
	} `yaml:"source"`
	Target struct {
		Type       string `yaml:"type"`
		BufferSize int    `yaml:"buffer_size"`
		MaxThreads int    `yaml:"max_threads"`
	} `yaml:"target"`
}

var ValidHashAlgorithms = map[string]bool{
	"md5":    true,
	"sha1":   true,
	"sha256": true,
}

var (
	ValidStorageTypes = map[string]bool{
		"hdd":     true,
		"ssd":     true,
		"network": true,
	}
	MaxThreadsLimit = 16
)

func (c *Config) Validate() error {
	if c.Source == "" {
		return fmt.Errorf("source directory is required")
	}
	if c.Target == "" {
		return fmt.Errorf("target directory is required")
	}
	if _, err := os.Stat(c.Source); err != nil {
		return fmt.Errorf("source directory invalid: %w", err)
	}
	if !ValidHashAlgorithms[c.Comparison.Algorithm] {
		return fmt.Errorf("invalid hash algorithm: %s, must be one of: md5, sha1, sha256", c.Comparison.Algorithm)
	}

	// Validate source storage config
	if !ValidStorageTypes[c.Storage.Source.Type] {
		return fmt.Errorf("invalid source device type: %s, must be one of: hdd, ssd, network", c.Storage.Source.Type)
	}
	if c.Storage.Source.MaxThreads < 0 || c.Storage.Source.MaxThreads > MaxThreadsLimit {
		return fmt.Errorf("source max threads must be between 0 and %d", MaxThreadsLimit)
	}

	// Validate target storage config
	if !ValidStorageTypes[c.Storage.Target.Type] {
		return fmt.Errorf("invalid target device type: %s, must be one of: hdd, ssd, network", c.Storage.Target.Type)
	}
	if c.Storage.Target.MaxThreads < 0 || c.Storage.Target.MaxThreads > MaxThreadsLimit {
		return fmt.Errorf("target max threads must be between 0 and %d", MaxThreadsLimit)
	}

	return nil
}

func (c *Config) SetDefaults() {
	if c.Comparison.Algorithm == "" {
		c.Comparison.Algorithm = "sha256"
	}
	if c.Comparison.Level == "" {
		c.Comparison.Level = "standard"
	}

	// Set source storage defaults
	if c.Storage.Source.Type == "" {
		c.Storage.Source.Type = "hdd"
	}
	if c.Storage.Source.MaxThreads == 0 {
		c.Storage.Source.MaxThreads = 4
	}

	// Set target storage defaults
	if c.Storage.Target.Type == "" {
		c.Storage.Target.Type = "hdd"
	}
	if c.Storage.Target.MaxThreads == 0 {
		c.Storage.Target.MaxThreads = 4
	}
}
