// internal/core/task/manager.go
package task

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

const DefaultBufferSize = 32 * 1024

type Manager struct {
	storage backup.StoragePort
	metrics backup.MetricsPort
	config  *backup.BackupConfig
}

func NewManager(storage backup.StoragePort, metrics backup.MetricsPort) *Manager {
	return &Manager{
		storage: storage,
		metrics: metrics,
	}
}

func (m *Manager) SetConfig(config *backup.BackupConfig) {
	m.config = config
}

func (m *Manager) CreateTasks(config *backup.BackupConfig) ([]backup.BackupTask, int, error) {
	m.SetConfig(config)
	var tasks []backup.BackupTask
	totalFiles := 0

	for _, folder := range config.FoldersToBackup {
		srcPath := filepath.Join(config.SourceDirectory, folder)
		dstPath := filepath.Join(config.TargetDirectory, folder)

		err := m.scanFolder(srcPath, dstPath, config.ExcludePatterns, &tasks, &totalFiles)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan folder %s: %w", folder, err)
		}
	}

	return tasks, totalFiles, nil
}

func (m *Manager) scanFolder(srcPath, dstPath string, excludePatterns []string, tasks *[]backup.BackupTask, totalFiles *int) error {
	return filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if m.isExcluded(info.Name(), excludePatterns) {
			return nil
		}

		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return fmt.Errorf("failed to determine relative path: %w", err)
		}

		destPath := filepath.Join(dstPath, relPath)
		*tasks = append(*tasks, backup.BackupTask{
			Source:      path,
			Destination: destPath,
			Size:        info.Size(),
			ModTime:     info.ModTime(),
		})
		*totalFiles++

		return nil
	})
}

func (m *Manager) isExcluded(filename string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return true
		}
	}
	return false
}

func (m *Manager) ShouldSkipFile(task backup.BackupTask) (bool, error) {
	exists, err := m.storage.Exists(task.Destination)
	if err != nil {
		return false, fmt.Errorf("failed to check destination existence: %w", err)
	}

	if !exists {
		return false, nil
	}

	srcMeta, err := m.storage.GetMetadata(task.Source)
	if err != nil {
		return false, fmt.Errorf("failed to get source metadata: %w", err)
	}

	dstMeta, err := m.storage.GetMetadata(task.Destination)
	if err != nil {
		return false, fmt.Errorf("failed to get destination metadata: %w", err)
	}

	if srcMeta.Size != dstMeta.Size {
		return false, nil
	}

	if m.config.DeepDuplicateCheck {
		return m.compareChecksums(task)
	}

	return true, nil
}

func (m *Manager) compareChecksums(task backup.BackupTask) (bool, error) {
	srcChecksum, err := m.storage.CalculateChecksum(task.Source)
	if err != nil {
		return false, fmt.Errorf("failed to calculate source checksum: %w", err)
	}

	dstChecksum, err := m.storage.CalculateChecksum(task.Destination)
	if err != nil {
		return false, fmt.Errorf("failed to calculate destination checksum: %w", err)
	}

	return srcChecksum == dstChecksum, nil
}

func (m *Manager) ExecuteTask(task backup.BackupTask) error {
	if err := m.storage.CreateDirectory(filepath.Dir(task.Destination)); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if _, err := m.storage.Copy(task.Source, task.Destination, DefaultBufferSize); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	if m.config.DeepDuplicateCheck {
		if equal, err := m.compareChecksums(task); err != nil {
			return fmt.Errorf("failed to verify copy: %w", err)
		} else if !equal {
			return fmt.Errorf("checksum mismatch for file: %s", task.Source)
		}
	}

	return nil
}

func (m *Manager) ValidateTask(task backup.BackupTask) error {
	exists, err := m.storage.Exists(task.Source)
	if err != nil {
		return fmt.Errorf("failed to check source existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("source file does not exist: %s", task.Source)
	}

	isDir, err := m.storage.IsDirectory(task.Source)
	if err != nil {
		return fmt.Errorf("failed to check if source is directory: %w", err)
	}
	if isDir {
		return fmt.Errorf("source is a directory: %s", task.Source)
	}

	return nil
}
