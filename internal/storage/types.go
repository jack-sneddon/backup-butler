// internal/storage/types.go
package storage

import (
	"time"

	"github.com/jack-sneddon/backup-butler/internal/types"
)

const (
	QuickHashSize = 64 * 1024 // 64KB for quick hash check
)

type ComparisonStrategy int

const (
	CompareMetadataOnly ComparisonStrategy = iota
	CompareQuickHash
	CompareFullChecksum
)

// ChecksumCalculator handles file checksum operations
type ChecksumCalculator struct{}

// CompareResult represents the result of comparing two files
type CompareResult struct {
	NeedsCopy   bool               `json:"needs_copy"`
	Reason      string             `json:"reason"`
	Strategy    ComparisonStrategy `json:"strategy"`
	Source      types.FileMetadata `json:"source"`
	Target      types.FileMetadata `json:"target"`
	QuickHash   string             `json:"quick_hash"`
	LastVersion string             `json:"last_version"`
	Warnings    []string           `json:"warnings,omitempty"` // Added field for corruption warnings
}

// CopyResult represents the result of a file copy operation
type CopyResult struct {
	Source      string        `json:"source"`
	Destination string        `json:"destination"`
	BytesCopied int64         `json:"bytes_copied"`
	Duration    time.Duration `json:"duration"`
	Error       error         `json:"error,omitempty"`
}
