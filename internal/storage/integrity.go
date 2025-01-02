// internal/storage/integrity.go
package storage

import (
	"fmt"

	"github.com/jack-sneddon/backup-butler/internal/types"
)

type IntegrityCheck struct {
	Path     string   `json:"path"`
	Issues   []string `json:"issues"`
	Severity string   `json:"severity"` // "warning" or "critical"
	Details  string   `json:"details"`  // Technical details for investigation
}

// Isolated integrity checking logic
func checkFileIntegrity(src string, srcInfo types.FileMetadata, lastVersion *types.FileVersionInfo) *IntegrityCheck {
	if lastVersion == nil {
		return nil // No previous version to check against
	}

	// Only check if modification time hasn't changed
	if !srcInfo.ModTime.Equal(lastVersion.ModTime) {
		return nil
	}

	var issues []string
	if srcInfo.Size != lastVersion.Size {
		issues = append(issues, "size changed without modification")
	}
	if srcInfo.Checksum != lastVersion.Checksum {
		issues = append(issues, "content changed without modification")
	}

	if len(issues) > 0 {
		return &IntegrityCheck{
			Path:     src,
			Issues:   issues,
			Severity: "warning",
			Details:  fmt.Sprintf("Previous backup: %s", lastVersion.ID),
		}
	}

	return nil
}
