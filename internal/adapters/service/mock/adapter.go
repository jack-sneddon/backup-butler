package mock

import (
	"context"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type MockBackupService struct {
	BackupFunc           func(ctx context.Context) error
	DryRunFunc           func(ctx context.Context) error
	GetVersionsFunc      func() []backup.BackupVersion
	GetVersionFunc       func(id string) (*backup.BackupVersion, error)
	GetLatestVersionFunc func() (*backup.BackupVersion, error)
}

func NewMockBackupService() *MockBackupService {
	return &MockBackupService{
		BackupFunc: func(ctx context.Context) error {
			return nil
		},
		DryRunFunc: func(ctx context.Context) error {
			return nil
		},
		GetVersionsFunc: func() []backup.BackupVersion {
			return nil
		},
		GetVersionFunc: func(id string) (*backup.BackupVersion, error) {
			return nil, nil
		},
		GetLatestVersionFunc: func() (*backup.BackupVersion, error) {
			return nil, nil
		},
	}
}

func (m *MockBackupService) Backup(ctx context.Context) error {
	return m.BackupFunc(ctx)
}

func (m *MockBackupService) DryRun(ctx context.Context) error {
	return m.DryRunFunc(ctx)
}

func (m *MockBackupService) GetVersions() []backup.BackupVersion {
	return m.GetVersionsFunc()
}

func (m *MockBackupService) GetVersion(id string) (*backup.BackupVersion, error) {
	return m.GetVersionFunc(id)
}

func (m *MockBackupService) GetLatestVersion() (*backup.BackupVersion, error) {
	return m.GetLatestVersionFunc()
}
