// internal/validation/core.go
package validation

import (
	"errors"
	"fmt"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/logger"
	"github.com/jack-sneddon/backup-butler/internal/scan"
	"github.com/jack-sneddon/backup-butler/internal/types"
)

// CriticalPath defines special validation requirements for specific paths
type CriticalPath struct {
	Path  string `yaml:"path"`
	Level string `yaml:"level"`
}

// Error definitions
var (
	ErrInvalidComparisonLevel = errors.New("invalid comparison level")
	ErrInvalidHashAlgorithm   = errors.New("invalid hash algorithm")
	ErrInvalidBufferSize      = errors.New("invalid buffer size")
)

// Validation constants
const (
	MinBufferSize = 4096
	MaxBufferSize = 10485760
)

// Supported hash algorithms
var SupportedAlgorithms = map[string]bool{
	"md5":    true,
	"sha1":   true,
	"sha256": true,
}

// ComparisonStrategy defines how files should be compared
type ComparisonStrategy interface {
	// Compare checks equality between source and target files
	Compare(source, target *scan.FileInfo) ComparisonResult
	// Level returns the comparison level used
	Level() types.ValidationLevel
}

// ValidatorOptions contains configuration for validators
type ValidatorOptions struct {
	BufferSize int    // Size of read buffer for content validation
	Algorithm  string // Hash algorithm (e.g., "sha256")
}

// ComparisonResult contains the outcome of a comparison operation
type ComparisonResult struct {
	Equal     bool
	Reason    string
	TimeTaken time.Duration
	BytesRead int64
}

// ValidationResult combines comparison results with validation status
type ValidationResult struct {
	Comparison  ComparisonResult
	RulesPassed bool
	Level       types.ValidationLevel
	Messages    []string
}

// FileValidator combines comparison and validation logic
type FileValidator struct {
	strategy ComparisonStrategy
	rules    ValidationRules
	stats    *ValidationStats
}

// ValidationRules defines integrity requirements
type ValidationRules struct {
	CriticalPaths []CriticalPathRule
	ScheduledDeep *ScheduledValidation
}

type CriticalPathRule struct {
	Pattern string
	Level   types.ValidationLevel
}

type ScheduledValidation struct {
	Enabled   bool
	Frequency string
	LastRun   time.Time
	Paths     []string
	Exclude   []string
}

// ValidationStats tracks validation metrics
type ValidationStats struct {
	QuickChecks    int
	StandardChecks int
	DeepChecks     int
	StartTime      time.Time
	EndTime        time.Time
}

func (opts *ValidatorOptions) Validate() error {
	if opts.BufferSize < MinBufferSize || opts.BufferSize > MaxBufferSize {
		return fmt.Errorf("%w: buffer size must be between %d and %d bytes",
			ErrInvalidBufferSize, MinBufferSize, MaxBufferSize)
	}

	if !SupportedAlgorithms[opts.Algorithm] {
		return fmt.Errorf("%w: must be one of: md5, sha1, sha256",
			ErrInvalidHashAlgorithm)
	}

	return nil
}

// ComparisonLevel indicates how thorough the comparison should be
//type ComparisonLevel string

// NewFileValidator creates a new validator with specified strategy and rules
func NewFileValidator(strategy ComparisonStrategy, rules ValidationRules) *FileValidator {
	return &FileValidator{
		strategy: strategy,
		rules:    rules,
		stats: &ValidationStats{
			StartTime: time.Now(),
		},
	}
}

// Validate performs both comparison and validation
func (v *FileValidator) Validate(source, target *scan.FileInfo) ValidationResult {
	valLogger := logger.WithGroup("validator").With(
		"source", source.Path,
		"target", target.Path,
		"size", source.Size,
		"strategy", string(v.strategy.Level()),
	)

	valLogger.Debug("Starting file validation")

	// Determine appropriate comparison level based on rules
	level := v.determineComparisonLevel(source.Path)

	// If we need a different level than our current strategy, create it
	if level != v.strategy.Level() {
		valLogger.Info("Validation level escalated",
			"from", v.strategy.Level(),
			"to", level,
		)
		v.strategy = NewStrategy(level, nil) // Use default options
	}

	// Perform comparison
	result := v.strategy.Compare(source, target)
	valLogger.Debug("Validation complete",
		"equal", result.Equal,
		"reason", result.Reason,
		"bytesRead", result.BytesRead,
		"timeTaken", result.TimeTaken,
	)

	// Track statistics
	switch level {
	case types.Quick:
		v.stats.QuickChecks++
	case types.Standard:
		v.stats.StandardChecks++
	case types.Deep:
		v.stats.DeepChecks++
	}

	// Validate against rules
	messages := v.validateRules(source, result)

	return ValidationResult{
		Comparison:  result,
		RulesPassed: len(messages) == 0,
		Level:       level,
		Messages:    messages,
	}
}

func (v *FileValidator) determineComparisonLevel(path string) types.ValidationLevel {
	// Check scheduled deep validation
	if v.shouldPerformScheduledDeep(path) {
		return types.Deep
	}

	// Use strategy's default level
	return v.strategy.Level()
}

// shouldPerformScheduledDeep checks if the path needs scheduled deep validation
func (v *FileValidator) shouldPerformScheduledDeep(path string) bool {
	if v.rules.ScheduledDeep == nil || !v.rules.ScheduledDeep.Enabled {
		return false
	}

	// Check if it's time for deep validation based on frequency and last run
	// Implementation depends on frequency format (daily, weekly, monthly)
	// For now, return false as placeholder
	return false
}

// getCriticalPathLevel checks if path matches any critical path patterns
func (v *FileValidator) getCriticalPathLevel(path string) types.ValidationLevel {
	// Implementation would use path matching against rules.CriticalPaths
	// For now, return empty as placeholder
	return ""
}

// validateRules checks if the comparison result satisfies all validation rules
func (v *FileValidator) validateRules(source *scan.FileInfo, result ComparisonResult) []string {
	var messages []string
	// Implementation would check various rules and collect validation messages
	return messages
}

// GetStats returns the current validation statistics
func (v *FileValidator) GetStats() ValidationStats {
	v.stats.EndTime = time.Now()
	return *v.stats
}

// NewStrategy creates a comparison strategy for the specified level
func NewStrategy(level types.ValidationLevel, opts *ValidatorOptions) ComparisonStrategy {
	switch level {
	case types.Quick:
		return NewQuickValidator()
	case types.Standard:
		return NewStandardValidator(opts)
	case types.Deep:
		return NewDeepValidator(opts)
	default:
		panic(fmt.Sprintf("unsupported comparison level: %s", level))
	}
}
