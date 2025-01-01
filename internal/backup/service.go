// internal/backup/service.go
package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/jack-sneddon/backup-butler/internal/config"
	"github.com/jack-sneddon/backup-butler/internal/storage"
	"github.com/jack-sneddon/backup-butler/internal/version"
)

type Service struct {
	config         *config.Config
	storageManager *storage.Manager
	copier         *storage.Copier
	versionMgr     *version.Manager
	stats          *atomic.Value // *BackupStats
}

type BackupStats struct {
	FilesProcessed int64
	BytesProcessed int64
	FilesCopied    int64
	BytesCopied    int64
	FilesSkipped   int64
	BytesSkipped   int64
	FilesFailed    int64
}

func NewService(cfg *config.Config) (*Service, error) {
	storageManager := storage.NewManager(cfg.BufferSize)
	copier := storage.NewCopier(storageManager, cfg.BufferSize)
	versionMgr, err := version.NewManager(cfg.TargetDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize version manager: %w", err)
	}

	stats := &atomic.Value{}
	stats.Store(&BackupStats{})

	return &Service{
		config:         cfg,
		storageManager: storageManager,
		copier:         copier,
		versionMgr:     versionMgr,
		stats:          stats,
	}, nil
}

func (s *Service) Backup(ctx context.Context) error {
	ver := s.versionMgr.StartNewVersion(s.config)
	fmt.Printf("Starting backup version: %s\n", ver.ID)

	tasks, err := s.scanSourceDirectory()
	if err != nil {
		s.versionMgr.CompleteVersion("failed")
		return fmt.Errorf("scan failed: %w", err)
	}

	fmt.Printf("Found %d files to process\n", len(tasks))
	if err := s.processBackupTasks(ctx, tasks); err != nil {
		s.versionMgr.CompleteVersion("failed")
		return fmt.Errorf("backup failed: %w", err)
	}

	stats := s.stats.Load().(*BackupStats)
	fmt.Printf("\nBackup completed:\n")
	fmt.Printf("Files processed: %d\n", stats.FilesProcessed)
	fmt.Printf("Files copied: %d (%.2f MB)\n", stats.FilesCopied, float64(stats.BytesCopied)/(1024*1024))
	fmt.Printf("Files skipped: %d (%.2f MB)\n", stats.FilesSkipped, float64(stats.BytesSkipped)/(1024*1024))
	fmt.Printf("Files failed: %d\n", stats.FilesFailed)

	return s.versionMgr.CompleteVersion("completed")
}

type BackupTask struct {
	SourcePath   string
	DestPath     string
	RelativePath string
}

func (s *Service) scanSourceDirectory() ([]BackupTask, error) {
	var tasks []BackupTask

	for _, folder := range s.config.FoldersToBackup {
		srcPath := filepath.Join(s.config.SourceDirectory, folder)

		err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				return nil
			}

			// Check exclude patterns
			for _, pattern := range s.config.ExcludePatterns {
				if matched, _ := filepath.Match(pattern, info.Name()); matched {
					return nil
				}
			}

			relPath, err := filepath.Rel(s.config.SourceDirectory, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}

			destPath := filepath.Join(s.config.TargetDirectory, relPath)

			tasks = append(tasks, BackupTask{
				SourcePath:   path,
				DestPath:     destPath,
				RelativePath: relPath,
			})

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to scan folder %s: %w", folder, err)
		}
	}

	return tasks, nil
}

func (s *Service) processBackupTasks(ctx context.Context, tasks []BackupTask) error {
	semaphore := make(chan struct{}, s.config.Concurrency)
	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	go func() {
		var wg sync.WaitGroup

		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return
			case semaphore <- struct{}{}:
				wg.Add(1)
				go func(t BackupTask) {
					defer wg.Done()
					defer func() { <-semaphore }()

					if err := s.processTask(ctx, t); err != nil {
						select {
						case errChan <- err:
						default:
						}
					}
				}(task)
			}
		}

		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	case <-doneChan:
		return nil
	}
}

func (s *Service) processTask(ctx context.Context, task BackupTask) error {
	meta, err := s.storageManager.GetMetadata(task.SourcePath)
	if err != nil {
		s.incrementStats(func(stats *BackupStats) {
			stats.FilesProcessed++
			stats.FilesFailed++
		})
		return fmt.Errorf("failed to get source metadata: %w", err)
	}

	// Pass the version manager to Compare
	compareResult, err := s.storageManager.Compare(task.SourcePath, task.DestPath, s.versionMgr)
	if err != nil {
		compareResult = storage.CompareResult{NeedsCopy: true, Reason: "comparison failed"}
	}

	if compareResult.NeedsCopy {
		result, err := s.copier.Copy(ctx, task.SourcePath, task.DestPath)
		if err != nil {
			s.incrementStats(func(stats *BackupStats) {
				stats.FilesProcessed++
				stats.FilesFailed++
			})
			return fmt.Errorf("failed to copy file: %w", err)
		}

		if err := s.copier.VerifyCopy(task.SourcePath, task.DestPath); err != nil {
			s.incrementStats(func(stats *BackupStats) {
				stats.FilesProcessed++
				stats.FilesFailed++
			})
			return fmt.Errorf("copy verification failed: %w", err)
		}

		s.incrementStats(func(stats *BackupStats) {
			stats.FilesProcessed++
			stats.FilesCopied++
			stats.BytesCopied += result.BytesCopied
			stats.BytesProcessed += result.BytesCopied
		})

		s.versionMgr.RecordFile(task.RelativePath, version.FileResult{
			Path:         meta.Path,
			Size:         meta.Size,
			ModTime:      meta.ModTime,
			Checksum:     meta.Checksum,
			Status:       "copied",
			CopyDuration: result.Duration,
			Metadata:     meta,
		})
	} else {
		s.incrementStats(func(stats *BackupStats) {
			stats.FilesProcessed++
			stats.FilesSkipped++
			stats.BytesSkipped += meta.Size
			stats.BytesProcessed += meta.Size
		})

		s.versionMgr.RecordFile(task.RelativePath, version.FileResult{
			Path:     meta.Path,
			Size:     meta.Size,
			ModTime:  meta.ModTime,
			Checksum: meta.Checksum,
			Status:   "skipped",
			Metadata: meta,
		})
	}

	return nil
}

func (s *Service) incrementStats(fn func(*BackupStats)) {
	for {
		oldStats := s.stats.Load().(*BackupStats)
		newStats := *oldStats
		fn(&newStats)
		if s.stats.CompareAndSwap(oldStats, &newStats) {
			return
		}
	}
}
