package validation

import (
	"errors"
	"fmt"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/scan"
)

// Error definitions
var (
	ErrInvalidComparisonLevel = errors.New("invalid comparison level")
	ErrInvalidHashAlgorithm   = errors.New("invalid hash algorithm")
	ErrInvalidBufferSize      = errors.New("invalid buffer size")
)

// Validation constants
const (
	MinBufferSize = 4096     // 4KB minimum
	MaxBufferSize = 10485760 // 10MB maximum
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
	Level() ComparisonLevel
}

// ValidatorOptions contains configuration for validators
type ValidatorOptions struct {
	BufferSize int    // Size of read buffer for content validation
	Algorithm  string // Hash algorithm (e.g., "sha256")
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
type ComparisonLevel string

const (
	Quick    ComparisonLevel = "quick"
	Standard ComparisonLevel = "standard"
	Deep     ComparisonLevel = "deep"
)

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
	Level       ComparisonLevel
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
	CriticalPaths []CriticalPath
	OnMismatch    ComparisonLevel
	ScheduledDeep *ScheduledValidation
}

// CriticalPath defines special validation requirements for specific paths
type CriticalPath struct {
	Pattern string
	Level   ComparisonLevel
}

// ScheduledValidation defines periodic deep validation requirements
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
	// Determine appropriate comparison level based on rules
	level := v.determineComparisonLevel(source.Path)

	// If we need a different level than our current strategy, create it
	if level != v.strategy.Level() {
		v.strategy = NewStrategy(level, nil) // Use default options
	}

	// Perform comparison
	result := v.strategy.Compare(source, target)

	// Track statistics
	switch level {
	case Quick:
		v.stats.QuickChecks++
	case Standard:
		v.stats.StandardChecks++
	case Deep:
		v.stats.DeepChecks++
	}

	// Check if we need to escalate validation level
	if !result.Equal && v.rules.OnMismatch > level {
		newStrategy := NewStrategy(v.rules.OnMismatch, nil)
		result = newStrategy.Compare(source, target)
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

func (v *FileValidator) determineComparisonLevel(path string) ComparisonLevel {
	// Check scheduled deep validation
	if v.shouldPerformScheduledDeep(path) {
		return Deep
	}

	// Check critical paths
	if level := v.getCriticalPathLevel(path); level != "" {
		return level
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
func (v *FileValidator) getCriticalPathLevel(path string) ComparisonLevel {
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
func NewStrategy(level ComparisonLevel, opts *ValidatorOptions) ComparisonStrategy {
	switch level {
	case Quick:
		return NewQuickValidator()
	case Standard:
		return NewStandardValidator(opts)
	case Deep:
		return NewDeepValidator(opts)
	default:
		panic(fmt.Sprintf("unsupported comparison level: %s", level))
	}
}
