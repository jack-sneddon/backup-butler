// internal/ui/shared/viewmodels/backup.go
package viewmodels

import "time"

type BackupProgress struct {
	CurrentFile     string
	CurrentPhase    string
	TotalFiles      int
	CompletedFiles  int
	SkippedFiles    int
	FailedFiles     int
	BytesProcessed  int64
	TotalBytes      int64
	TransferRate    float64
	TimeElapsed     time.Duration
	TimeRemaining   time.Duration
	PercentComplete float64
}

type BackupOperation struct {
	IsRunning    bool
	IsDryRun     bool
	CurrentPhase string // "scanning", "copying", "verifying", "completed"
	Progress     BackupProgress
	Errors       []string
}

type ConfigViewModel struct {
	SourcePath      string
	TargetPath      string
	SelectedFolders []string
	UseChecksum     bool
	Concurrency     int
	ShowProgress    bool
}

type VersionViewModel struct {
	ID           string
	TimeStamp    time.Time
	Duration     time.Duration
	FileCount    int
	Size         int64
	Status       string
	FilesCopied  int
	FilesSkipped int
	FilesFailed  int
}
