// internal/types/common.go
package types

import "time"

// FileMetadata represents file-level information
type FileMetadata struct {
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"mod_time"`
	Checksum string    `json:"checksum"`
}

// FileVersionInfo holds version information for a file
type FileVersionInfo struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"mod_time"`
	Checksum  string    `json:"checksum"`
	QuickHash string    `json:"quick_hash"`
}

// VersionManager defines the interface for version operations
type VersionManager interface {
	GetFileLastVersion(path string) (*FileVersionInfo, error)
}
