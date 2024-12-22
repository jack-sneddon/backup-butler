package mock

import (
	"fmt"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type MockVersionManager struct {
	versions   []backup.BackupVersion
	currentVer *backup.BackupVersion
}

func NewMockVersionManager() *MockVersionManager {
	return &MockVersionManager{
		versions: make([]backup.BackupVersion, 0),
	}
}

func (m *MockVersionManager) StartNewVersion(config *backup.BackupConfig) *backup.BackupVersion {
	version := &backup.BackupVersion{
		ID:         time.Now().Format("20060102-150405"),
		Timestamp:  time.Now(),
		Files:      make(map[string]backup.FileMetadata),
		Status:     "In Progress",
		ConfigUsed: *config,
	}
	m.currentVer = version
	return version
}

func (m *MockVersionManager) AddFile(path string, metadata backup.FileMetadata) {
	if m.currentVer != nil {
		m.currentVer.Files[path] = metadata
		m.currentVer.Size += metadata.Size
		m.currentVer.Stats.TotalFiles++
		m.currentVer.Stats.TotalBytes += metadata.Size
	}
}

func (m *MockVersionManager) CompleteVersion(stats backup.BackupStats) error {
	if m.currentVer == nil {
		return fmt.Errorf("no backup version in progress")
	}

	m.currentVer.Status = "Completed"
	m.currentVer.Duration = time.Since(m.currentVer.Timestamp)
	m.currentVer.Stats = stats

	m.versions = append(m.versions, *m.currentVer)
	m.currentVer = nil

	return nil
}

func (m *MockVersionManager) GetVersions() []backup.BackupVersion {
	return m.versions
}

func (m *MockVersionManager) GetVersion(id string) (*backup.BackupVersion, error) {
	for _, ver := range m.versions {
		if ver.ID == id {
			return &ver, nil
		}
	}
	return nil, fmt.Errorf("version not found: %s", id)
}

func (m *MockVersionManager) GetLatestVersion() (*backup.BackupVersion, error) {
	if len(m.versions) == 0 {
		return nil, fmt.Errorf("no backup versions found")
	}
	return &m.versions[len(m.versions)-1], nil
}
