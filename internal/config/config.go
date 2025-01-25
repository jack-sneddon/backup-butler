// internal/config/config.go
package config

import (
	"fmt"
	"os"
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

	Storage struct {
		DeviceType string `yaml:"device_type"`
		MaxThreads int    `yaml:"max_threads"`
	} `yaml:"storage"`
}

var ValidHashAlgorithms = map[string]bool{
	"md5":    true,
	"sha1":   true,
	"sha256": true,
}

var (
	ValidDeviceTypes = map[string]bool{
		"hdd":   true,
		"ssd":   true,
		"cloud": true,
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
	if !ValidDeviceTypes[c.Storage.DeviceType] {
		return fmt.Errorf("invalid device type: %s, must be one of: hdd, ssd, cloud", c.Storage.DeviceType)
	}
	if c.Storage.MaxThreads == 0 {
		return fmt.Errorf("max threads is required")
	}
	if c.Storage.MaxThreads < 1 || c.Storage.MaxThreads > MaxThreadsLimit {
		return fmt.Errorf("max threads must be between 1 and %d", MaxThreadsLimit)
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
	if c.Storage.DeviceType == "" {
		c.Storage.DeviceType = "hdd"
	}
	if c.Storage.MaxThreads <= 0 {
		c.Storage.MaxThreads = 4
	}
}
