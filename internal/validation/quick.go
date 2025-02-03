// internal/validation/quick.go
package validation

import (
	"time"

	"github.com/jack-sneddon/backup-butler/internal/scan"
)

// QuickValidator implements metadata-only comparison
type QuickValidator struct{}

func NewQuickValidator() *QuickValidator {
	return &QuickValidator{}
}

func (v *QuickValidator) Level() ValidationLevel {
	return Quick
}

// Compare checks equality between source and target files using metadata only
func (v *QuickValidator) Compare(source, target *scan.FileInfo) ComparisonResult {
	start := time.Now()

	// Basic size comparison
	if source.Size != target.Size {
		return ComparisonResult{
			Equal:     false,
			Reason:    "File sizes differ",
			TimeTaken: time.Since(start),
		}
	}

	// ModTime comparison with tolerance
	// Allow 2-second tolerance for filesystem timestamp differences
	const modTimeToleranceSeconds = 2
	modTimeDiff := abs(source.ModTime - target.ModTime)
	if modTimeDiff > modTimeToleranceSeconds {
		return ComparisonResult{
			Equal:     false,
			Reason:    "Modification times differ significantly",
			TimeTaken: time.Since(start),
		}
	}

	return ComparisonResult{
		Equal:     true,
		Reason:    "Metadata match",
		TimeTaken: time.Since(start),
	}
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
