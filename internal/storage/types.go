// internal/storage/types.go
package storage

import (
	"time"
)

// CopyResult represents the result of a file copy operation
type CopyResult struct {
	Source      string        `json:"source"`
	Destination string        `json:"destination"`
	BytesCopied int64         `json:"bytes_copied"`
	Duration    time.Duration `json:"duration"`
	Error       error         `json:"error,omitempty"`
}
