// internal/core/backup/service.go
package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type Service struct {
	config     *backup.BackupConfig
	storage    backup.StoragePort
	metrics    backup.MetricsPort
	versioner  backup.VersionManagerPort
	taskMgr    backup.TaskManagerPort
	workerPool backup.WorkerPoolPort
}

func NewService(
	config *backup.BackupConfig,
	storage backup.StoragePort,
	metrics backup.MetricsPort,
	versioner backup.VersionManagerPort,
	taskMgr backup.TaskManagerPort,
	workerPool backup.WorkerPoolPort,
) *Service {
	return &Service{
		config:     config,
		storage:    storage,
		metrics:    metrics,
		versioner:  versioner,
		taskMgr:    taskMgr,
		workerPool: workerPool,
	}
}

func (s *Service) Backup(ctx context.Context) error {
	// Create backup tasks
	tasks, totalFiles, err := s.taskMgr.CreateTasks(s.config)
	if err != nil {
		return fmt.Errorf("failed to create backup tasks: %w", err)
	}

	if !s.config.Options.Quiet {
		fmt.Printf("Starting backup of %d files...\n", totalFiles)
	}

	// Initialize metrics
	if metricsSetter, ok := s.metrics.(interface{ SetTotalFiles(int) }); ok {
		metricsSetter.SetTotalFiles(totalFiles)
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

func (s *Service) DryRun(ctx context.Context) error {
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

	stats := &DryRunStats{TotalFiles: totalFiles}

	// Process each task sequentially - pure analysis, no directory creation
	for _, task := range tasks {
		exists, _ := s.storage.Exists(task.Destination)
		if exists {
			shouldSkip, _ := s.taskMgr.ShouldSkipFile(task)
			if shouldSkip {
				stats.UpdateForFile(task, false)
				stats.LogFileStatus(task, false, s.config.Options.Quiet)
				continue
			}
		}

		stats.UpdateForFile(task, true)
		stats.LogFileStatus(task, true, s.config.Options.Quiet)
	}

	stats.Print()
	return nil
}

func (s *Service) GetVersions() []backup.BackupVersion {
	return s.versioner.GetVersions()
}

func (s *Service) GetVersion(id string) (*backup.BackupVersion, error) {
	return s.versioner.GetVersion(id)
}

func (s *Service) GetLatestVersion() (*backup.BackupVersion, error) {
	return s.versioner.GetLatestVersion()
}
