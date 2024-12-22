package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/adapters/metrics/collector"
	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type BackupServiceImpl struct {
	config     *backup.BackupConfig
	storage    backup.StoragePort
	metrics    backup.MetricsPort
	versioner  backup.VersionManagerPort
	taskMgr    backup.TaskManagerPort
	workerPool backup.WorkerPoolPort
}

func NewBackupService(
	config *backup.BackupConfig,
	storage backup.StoragePort,
	metrics backup.MetricsPort,
	versioner backup.VersionManagerPort,
	taskMgr backup.TaskManagerPort,
	workerPool backup.WorkerPoolPort,
) *BackupServiceImpl {
	return &BackupServiceImpl{
		config:     config,
		storage:    storage,
		metrics:    metrics,
		versioner:  versioner,
		taskMgr:    taskMgr,
		workerPool: workerPool,
	}
}

func (s *BackupServiceImpl) Backup(ctx context.Context) error {
	// Create backup tasks
	tasks, totalFiles, err := s.taskMgr.CreateTasks(s.config)
	if err != nil {
		return fmt.Errorf("failed to create backup tasks: %w", err)
	}

	if !s.config.Options.Quiet {
		fmt.Printf("Starting backup of %d files...\n", totalFiles)
	}

	// Initialize metrics
	if collector, ok := s.metrics.(*collector.MetricsCollector); ok {
		collector.SetTotalFiles(totalFiles)
	}
	s.metrics.StartTracking(ctx)

	// Start new backup version
	s.versioner.StartNewVersion(s.config)

	// Process tasks
	resultCh := s.workerPool.Execute(ctx, tasks)

	// Process results as they come in
	for result := range resultCh {
		// Update versioning info
		if result.Status != "failed" {
			metadata, err := s.storage.GetMetadata(result.Task.Source)
			if err != nil {
				fmt.Printf("Error: Failed to get metadata for %s: %v\n", result.Task.Source, err)
			} else {
				s.versioner.AddFile(result.Task.Source, metadata)
			}
		}

		// Update metrics based on result
		switch result.Status {
		case "completed":
			s.metrics.IncrementCompleted(result.Bytes)
		case "skipped":
			s.metrics.IncrementSkipped(result.Bytes)
		case "failed":
			s.metrics.IncrementFailed()
			if !s.config.Options.Quiet {
				fmt.Printf("Error: Failed to process %s: %v\n", result.Task.Source, result.Error)
			}
		}
	}

	// Get final stats and complete version
	stats := s.metrics.GetStats()
	if err := s.versioner.CompleteVersion(stats); err != nil {
		fmt.Printf("Error: Failed to save backup version: %v\n", err)
	}

	// Wait for any final updates
	time.Sleep(200 * time.Millisecond)

	// Display final summary
	s.metrics.DisplayFinalSummary()

	return nil
}

func (s *BackupServiceImpl) DryRun(ctx context.Context) error {
	// Store original config target and create tasks without creating directories
	originalTarget := s.config.TargetDirectory
	tempTarget := filepath.Join(os.TempDir(), "backup-butler-dryrun")
	s.config.TargetDirectory = tempTarget

	// Restore original target when done
	defer func() {
		s.config.TargetDirectory = originalTarget
	}()

	// Create backup tasks
	tasks, totalFiles, err := s.taskMgr.CreateTasks(s.config)
	if err != nil {
		return fmt.Errorf("failed to create backup tasks: %w", err)
	}

	fmt.Printf("Starting dry run analysis of %d files...\n", totalFiles)

	var (
		filesToCopy  = 0
		totalBytes   = int64(0)
		skippedFiles = 0
		skippedBytes = int64(0)
	)

	// Process each task sequentially - pure analysis, no directory creation
	for _, task := range tasks {
		exists, _ := s.storage.Exists(task.Destination)
		if exists {
			shouldSkip, _ := s.taskMgr.ShouldSkipFile(task)
			if shouldSkip {
				skippedFiles++
				skippedBytes += task.Size
				if !s.config.Options.Quiet {
					fmt.Printf("Would skip: %s (%.2f MB)\n",
						filepath.Base(task.Source),
						float64(task.Size)/1024/1024)
				}
				continue
			}
		}

		filesToCopy++
		totalBytes += task.Size
		if !s.config.Options.Quiet {
			fmt.Printf("Would copy: %s (%.2f MB)\n",
				filepath.Base(task.Source),
				float64(task.Size)/1024/1024)
		}
	}

	// Print summary
	fmt.Printf("\nDry Run Summary:\n")
	fmt.Printf("Files to copy:  %d (%.2f MB)\n", filesToCopy, float64(totalBytes)/1024/1024)
	fmt.Printf("Files to skip:  %d (%.2f MB)\n", skippedFiles, float64(skippedBytes)/1024/1024)
	fmt.Printf("Total files:    %d\n", totalFiles)
	fmt.Printf("Total size:     %.2f MB\n", float64(totalBytes+skippedBytes)/1024/1024)

	return nil
}

func (s *BackupServiceImpl) GetVersions() []backup.BackupVersion {
	return s.versioner.GetVersions()
}

func (s *BackupServiceImpl) GetVersion(id string) (*backup.BackupVersion, error) {
	return s.versioner.GetVersion(id)
}

func (s *BackupServiceImpl) GetLatestVersion() (*backup.BackupVersion, error) {
	return s.versioner.GetLatestVersion()
}
