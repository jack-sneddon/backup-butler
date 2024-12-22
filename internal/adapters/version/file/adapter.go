package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type FileVersionManager struct {
	baseDir    string
	versions   []backup.BackupVersion
	currentVer *backup.BackupVersion
}

func NewFileVersionManager(baseDir string) (*FileVersionManager, error) {
	vm := &FileVersionManager{
		baseDir: baseDir,
	}

	// Create versions directory if it doesn't exist
	versionsDir := filepath.Join(baseDir, ".versions")
	if err := os.MkdirAll(versionsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create versions directory: %w", err)
	}

	// Load existing versions
	if err := vm.loadVersions(); err != nil {
		return nil, err
	}

	return vm, nil
}

func (vm *FileVersionManager) StartNewVersion(config *backup.BackupConfig) *backup.BackupVersion {
	version := &backup.BackupVersion{
		ID:         time.Now().Format("20060102-150405"),
		Timestamp:  time.Now(),
		Files:      make(map[string]backup.FileMetadata),
		Status:     "In Progress",
		ConfigUsed: *config,
	}
	vm.currentVer = version
	return version
}

func (vm *FileVersionManager) AddFile(path string, metadata backup.FileMetadata) {
	if vm.currentVer != nil {
		vm.currentVer.Files[path] = metadata
		vm.currentVer.Size += metadata.Size
		vm.currentVer.Stats.TotalFiles++
		vm.currentVer.Stats.TotalBytes += metadata.Size
	}
}

func (vm *FileVersionManager) CompleteVersion(stats backup.BackupStats) error {
	if vm.currentVer == nil {
		return fmt.Errorf("no backup version in progress")
	}

	vm.currentVer.Status = "Completed"
	vm.currentVer.Duration = time.Since(vm.currentVer.Timestamp)
	vm.currentVer.Stats = stats

	// Save the version
	if err := vm.saveVersion(vm.currentVer); err != nil {
		return err
	}

	vm.versions = append(vm.versions, *vm.currentVer)
	vm.currentVer = nil

	return nil
}

func (vm *FileVersionManager) saveVersion(ver *backup.BackupVersion) error {
	filename := filepath.Join(vm.baseDir, ".versions", ver.ID+".json")

	data, err := json.MarshalIndent(ver, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal version data: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to save version file: %w", err)
	}

	return nil
}

func (vm *FileVersionManager) loadVersions() error {
	versionsDir := filepath.Join(vm.baseDir, ".versions")
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read versions directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			filename := filepath.Join(versionsDir, entry.Name())
			data, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("failed to read version file %s: %w", entry.Name(), err)
			}

			var version backup.BackupVersion
			if err := json.Unmarshal(data, &version); err != nil {
				return fmt.Errorf("failed to parse version file %s: %w", entry.Name(), err)
			}

			vm.versions = append(vm.versions, version)
		}
	}

	return nil
}

func (vm *FileVersionManager) GetVersions() []backup.BackupVersion {
	return vm.versions
}

func (vm *FileVersionManager) GetVersion(id string) (*backup.BackupVersion, error) {
	for _, ver := range vm.versions {
		if ver.ID == id {
			return &ver, nil
		}
	}
	return nil, fmt.Errorf("version not found: %s", id)
}

func (vm *FileVersionManager) GetLatestVersion() (*backup.BackupVersion, error) {
	if len(vm.versions) == 0 {
		return nil, fmt.Errorf("no backup versions found")
	}
	return &vm.versions[len(vm.versions)-1], nil
}
