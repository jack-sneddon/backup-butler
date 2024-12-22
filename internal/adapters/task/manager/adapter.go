package manager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type TaskManager struct {
	storage backup.StoragePort
	metrics backup.MetricsPort
	//logger  backup.LoggerPort
	config *backup.BackupConfig
}

func NewTaskManager(
	storage backup.StoragePort,
	metrics backup.MetricsPort,
	//logger backup.LoggerPort,
) *TaskManager {
	return &TaskManager{
		storage: storage,
		metrics: metrics,
		//logger:  logger,
	}
}

func (t *TaskManager) SetConfig(config *backup.BackupConfig) {
	t.config = config
}

func (t *TaskManager) CreateTasks(config *backup.BackupConfig) ([]backup.BackupTask, int, error) {
	t.SetConfig(config)
	var tasks []backup.BackupTask
	totalFiles := 0

	// Process each folder
	for _, folder := range config.FoldersToBackup {
		srcPath := filepath.Join(config.SourceDirectory, folder)
		dstPath := filepath.Join(config.TargetDirectory, folder)

		//t.logger.Info("Scanning folder: %s", folder)
		err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				//t.logger.Error("Failed to access path %s: %v", path, err)
				return err
			}

			if info.IsDir() {
				return nil
			}

			// Check exclude patterns
			for _, pattern := range config.ExcludePatterns {
				if matched, _ := filepath.Match(pattern, info.Name()); matched {
					//t.logger.Debug("Excluding file: %s", path)
					return nil
				}
			}

			relPath, err := filepath.Rel(srcPath, path)
			if err != nil {
				return fmt.Errorf("failed to determine relative path: %w", err)
			}

			destPath := filepath.Join(dstPath, relPath)
			tasks = append(tasks, backup.BackupTask{
				Source:      path,
				Destination: destPath,
				Size:        info.Size(),
				ModTime:     info.ModTime(),
			})
			totalFiles++

			return nil
		})

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan folder %s: %w", folder, err)
		}
	}

	//t.logger.Info("Found %d files to process", totalFiles)
	return tasks, totalFiles, nil
}

func (t *TaskManager) ShouldSkipFile(task backup.BackupTask) (bool, error) {
	// Check if destination exists
	exists, err := t.storage.Exists(task.Destination)
	if err != nil {
		return false, fmt.Errorf("failed to check destination existence: %w", err)
	}

	if !exists {
		return false, nil
	}

	// Compare file sizes first
	srcMeta, err := t.storage.GetMetadata(task.Source)
	if err != nil {
		return false, fmt.Errorf("failed to get source metadata: %w", err)
	}

	dstMeta, err := t.storage.GetMetadata(task.Destination)
	if err != nil {
		return false, fmt.Errorf("failed to get destination metadata: %w", err)
	}

	if srcMeta.Size != dstMeta.Size {
		return false, nil
	}

	// If size matches and deep check is enabled, compare checksums
	if t.config.DeepDuplicateCheck {
		srcChecksum, err := t.storage.CalculateChecksum(task.Source)
		if err != nil {
			return false, fmt.Errorf("failed to calculate source checksum: %w", err)
		}

		dstChecksum, err := t.storage.CalculateChecksum(task.Destination)
		if err != nil {
			return false, fmt.Errorf("failed to calculate destination checksum: %w", err)
		}

		return srcChecksum == dstChecksum, nil
	}

	// If deep check is disabled, consider files identical if sizes match
	return true, nil
}

func (t *TaskManager) ExecuteTask(task backup.BackupTask) error {
	// Create destination directory if needed
	if err := t.storage.CreateDirectory(filepath.Dir(task.Destination)); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Copy the file
	if _, err := t.storage.Copy(task.Source, task.Destination, 32*1024); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Verify the copy if deep check is enabled
	if t.config.DeepDuplicateCheck {
		srcChecksum, err := t.storage.CalculateChecksum(task.Source)
		if err != nil {
			return fmt.Errorf("failed to calculate source checksum: %w", err)
		}

		dstChecksum, err := t.storage.CalculateChecksum(task.Destination)
		if err != nil {
			return fmt.Errorf("failed to calculate destination checksum: %w", err)
		}

		if srcChecksum != dstChecksum {
			return fmt.Errorf("checksum mismatch for file: %s", task.Source)
		}
	}

	return nil
}

func (t *TaskManager) ValidateTask(task backup.BackupTask) error {
	// Validate source exists
	exists, err := t.storage.Exists(task.Source)
	if err != nil {
		return fmt.Errorf("failed to check source existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("source file does not exist: %s", task.Source)
	}

	// Validate source is not a directory
	isDir, err := t.storage.IsDirectory(task.Source)
	if err != nil {
		return fmt.Errorf("failed to check if source is directory: %w", err)
	}
	if isDir {
		return fmt.Errorf("source is a directory: %s", task.Source)
	}

	return nil
}
