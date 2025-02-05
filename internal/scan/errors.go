// internal/scan/errors.go
package scan

import (
	"fmt"
	"path/filepath"

	"github.com/jack-sneddon/backup-butler/internal/logger"
)

// ScanError represents errors that can occur during scanning
type ScanError struct {
	Path string
	Op   string
	Err  error
}

func (e *ScanError) Error() string {
	return fmt.Sprintf("scan error: %s on path %s: %v", e.Op, e.Path, e.Err)
}

// NewScanError creates a new ScanError
func NewScanError(path string, op string, err error) *ScanError {
	return &ScanError{
		Path: path,
		Op:   op,
		Err:  err,
	}
}

// Scanner configuration options
type ScannerOptions struct {
	ExcludePatterns  []string
	IncludeFolders   []string
	MaxDepth         int
	BufferSize       int
	DefaultLevel     string
	ValidationConfig *ValidationConfig
}

type ValidationConfig struct {
	DefaultLevel  string `yaml:"default_level"`
	OnMismatch    string `yaml:"on_mismatch"`
	BufferSize    int    `yaml:"buffer_size"`
	HashAlgorithm string `yaml:"hash_algorithm"`
}

// Enhanced Progress tracking
type Progress struct {
	CurrentDir     string
	ScannedDirs    int
	ScannedFiles   int
	ProcessedBytes int64
	TotalBytes     int64
	ExcludedFiles  int
	ExcludedDirs   int
	Errors         []string
	Phase          string // "counting", "scanning", "comparing"
	TotalFiles     int
}

func (p *Progress) AddError(err error) {
	if err != nil {
		p.Errors = append(p.Errors, err.Error())
	}
}

// internal/scan/errors.go

// shouldIncludeFolder checks if a folder should be included based on config
func shouldIncludeFolder(path string, includeFolders []string) bool {
	// If no specific folders are specified, include all
	if len(includeFolders) == 0 {
		return true
	}

	// Get relative path from root directory
	basePath := filepath.Base(path)

	for _, folder := range includeFolders {
		// Direct match with base name
		if folder == basePath {
			return true
		}

		// Check if any parent directory matches
		dir := path
		for dir != "." && dir != "/" {
			if filepath.Base(dir) == folder {
				return true
			}
			dir = filepath.Dir(dir)
		}
	}
	return false
}

// matchesPattern checks if a path matches any of the exclude patterns
func matchesPattern(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}

	// Get the path relative to the scanning root
	/*
		logger.Get().Debugw("Checking pattern match",
			"path", path,
			"patterns", patterns)
	*/

	for _, pattern := range patterns {
		// Convert pattern into filepath-compatible format
		pattern = filepath.FromSlash(pattern)
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			logger.Get().Debugw("Pattern match error",
				"pattern", pattern,
				"path", path,
				"error", err)
			continue
		}
		if matched {
			/*
				logger.Get().Debugw("Pattern matched",
					"pattern", pattern,
					"path", path)
			*/
			return true
		}
	}
	return false
}
