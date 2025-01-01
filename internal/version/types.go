// internal/version/types.go
package version

import (
	"fmt"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/config"
	"github.com/jack-sneddon/backup-butler/internal/types"
)

type BackupVersion struct {
	ID        string                `json:"id"`
	StartTime time.Time             `json:"start_time"`
	EndTime   time.Time             `json:"end_time"`
	Status    string                `json:"status"` // "running", "completed", "failed"
	Config    *config.Config        `json:"config"` // Configuration used
	Stats     BackupStats           `json:"stats"`
	Files     map[string]FileResult `json:"files"` // Path -> Result
}

type BackupStats struct {
	TotalFiles   int   `json:"total_files"`
	FilesCopied  int   `json:"files_copied"`
	FilesSkipped int   `json:"files_skipped"`
	FilesFailed  int   `json:"files_failed"`
	BytesCopied  int64 `json:"bytes_copied"`
	BytesSkipped int64 `json:"bytes_skipped"`
}

type FileResult struct {
	Path         string             `json:"path"` // Relative path from source
	Size         int64              `json:"size"`
	ModTime      time.Time          `json:"mod_time"`
	Checksum     string             `json:"checksum"`
	Status       string             `json:"status"` // "copied", "skipped", "failed"
	CopyDuration time.Duration      `json:"copy_duration"`
	Error        string             `json:"error,omitempty"`
	Metadata     types.FileMetadata `json:"metadata"`
	QuickHash    string             `json:"quick_hash"` // Quick hash of first 64KB
}

// VersionSummary provides a condensed view of a backup version
type VersionSummary struct {
	ID           string    `json:"id"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Status       string    `json:"status"`
	TotalFiles   int       `json:"total_files"`
	TotalBytes   int64     `json:"total_bytes"`
	CopiedFiles  int       `json:"copied_files"`
	CopiedBytes  int64     `json:"copied_bytes"`
	SkippedFiles int       `json:"skipped_files"`
	SkippedBytes int64     `json:"skipped_bytes"`
	FailedFiles  int       `json:"failed_files"`
}

func FormatSize(bytes int64) string {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
