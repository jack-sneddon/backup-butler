// internal/progress/types.go
package progress

import "time"

type DirectoryProgress struct {
	Path      string
	Total     int64
	Processed int64
	Files     int
	Done      int
	StartTime time.Time
}

type Progress struct {
	Current     *DirectoryProgress
	TotalFiles  int
	TotalBytes  int64
	Processed   int
	BytesDone   int64
	StartTime   time.Time
	DisplayDone chan bool
	Phase       string
}

type Tracker interface {
	StartDirectory(path string, totalBytes int64, fileCount int) error
	UpdateProgress(bytes int64)
	FinishDirectory() error
	ScanDirectory(path string) error
	GetProgress() *Progress
	Start() error
	Stop() error
}
