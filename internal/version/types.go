// internal/version/types.go
package version

import (
	"time"
)

type FileIndex struct {
	LastUpdated time.Time               `json:"last_updated"`
	Files       map[string]FileMetadata `json:"files"`
}

type FileMetadata struct {
	LastBackupID string    `json:"last_backup_id"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"mod_time"`
	Checksum     string    `json:"checksum"`
}

type BackupVersion struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Stats     struct {
		Directories map[string]DirectoryStats `json:"directories"`
		Total       BackupStats               `json:"total"`
	} `json:"stats"`
	Changes []FileChange `json:"changes"`
}

type FileChange struct {
	Path      string    `json:"path"`
	Action    string    `json:"action"` // "copied", "skipped", "failed"
	Size      int64     `json:"size"`
	Timestamp time.Time `json:"timestamp"`
	Checksum  string    `json:"checksum,omitempty"` // Only for copied files
}

type DirectoryStats struct {
	TotalFiles   int   `json:"total_files"`
	TotalBytes   int64 `json:"total_bytes"`
	CopiedFiles  int   `json:"copied_files"`
	CopiedBytes  int64 `json:"copied_bytes"`
	SkippedFiles int   `json:"skipped_files"`
	SkippedBytes int64 `json:"skipped_bytes"`
	FailedFiles  int   `json:"failed_files"`
}

type BackupStats struct {
	TotalFiles   int   `json:"total_files"`
	FilesCopied  int   `json:"files_copied"`
	FilesSkipped int   `json:"files_skipped"`
	FilesFailed  int   `json:"files_failed"`
	BytesCopied  int64 `json:"bytes_copied"`
	BytesSkipped int64 `json:"bytes_skipped"`
}

// For command-line display
type VersionSummary struct {
	ID        string                    `json:"id"`
	StartTime time.Time                 `json:"start_time"`
	EndTime   time.Time                 `json:"end_time"`
	Stats     BackupStats               `json:"stats"`
	DirStats  map[string]DirectoryStats `json:"dir_stats"`
}
