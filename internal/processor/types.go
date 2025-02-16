// internal/processor/types.go
package processor

import (
	"time"

	"github.com/jack-sneddon/backup-butler/internal/progress"
)

type ProcessorOptions struct {
	PreserveMetadata bool
	BufferSize       int
	MaxThreads       int
	StorageType      string // "hdd", "ssd", "network"
	Progress         progress.Tracker
}

type DirectoryProcessor interface {
	ProcessDirectory(sourcePath, targetPath string) error
}

type DirectoryResult struct {
	FilesProcessed int
	BytesProcessed int64
	TimeElapsed    time.Duration
	FilesPerSecond float64
	BytesPerSecond int64
	Errors         []error
}
