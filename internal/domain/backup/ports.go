// internal/domain/backup/ports.go
package backup

import (
	"context"
)

// TaskResult represents the outcome of a backup task
type TaskResult struct {
	Task   BackupTask
	Status string // "completed", "skipped", "failed"
	Bytes  int64
	Error  error
}

// WorkerPoolPort defines the worker pool operations
type WorkerPoolPort interface {
	Execute(ctx context.Context, tasks []BackupTask) <-chan TaskResult
}

// MetricsPort defines metrics operations
type MetricsPort interface {
	StartTracking(ctx context.Context)
	IncrementCompleted(bytes int64)
	IncrementSkipped(bytes int64)
	IncrementFailed()
	GetStats() BackupStats
	DisplayProgress()
	DisplayFinalSummary()
}

// BackupService defines the core backup operations
type BackupService interface {
	// Core operations
	Backup(ctx context.Context) error
	DryRun(ctx context.Context) error

	// Version operations
	GetVersions() []BackupVersion
	GetVersion(id string) (*BackupVersion, error)
	GetLatestVersion() (*BackupVersion, error)
}

// StoragePort defines storage operations
type StoragePort interface {
	CalculateChecksum(filePath string) (string, error)
	Copy(src, dst string, bufferSize int) (int64, error)
	Exists(path string) (bool, error)
	GetMetadata(path string) (FileMetadata, error)
	CreateDirectory(path string) error
	IsDirectory(path string) (bool, error)
}

// VersionManagerPort handles backup versioning
type VersionManagerPort interface {
	StartNewVersion(config *BackupConfig) *BackupVersion
	AddFile(path string, metadata FileMetadata)
	CompleteVersion(stats BackupStats) error
	GetVersions() []BackupVersion
	GetVersion(id string) (*BackupVersion, error)
	GetLatestVersion() (*BackupVersion, error)
}

/*
// MonitoringPort defines monitoring operations
type MonitoringPort interface {
	LoggerPort
	MetricsPort
}

// LoggerPort defines logging operations
type LoggerPort interface {
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	Close() error
}
*/

// ConfigLoaderPort defines configuration loading operations
type ConfigLoaderPort interface {
	Load(path string) (*BackupConfig, error)
	Validate(config *BackupConfig) error
}

// TaskManagerPort handles backup task management
type TaskManagerPort interface {
	CreateTasks(config *BackupConfig) ([]BackupTask, int, error)
	ShouldSkipFile(task BackupTask) (bool, error)
	ExecuteTask(task BackupTask) error
	ValidateTask(task BackupTask) error
}
