// internal/adapters/service/backup/adapter.go
package backup

import (
	"context"

	corebackup "github.com/jack-sneddon/backup-butler/internal/core/backup"
	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type BackupServiceAdapter struct {
	service *corebackup.Service
}

func NewBackupService(
	config *backup.BackupConfig,
	storage backup.StoragePort,
	metrics backup.MetricsPort,
	versioner backup.VersionManagerPort,
	taskMgr backup.TaskManagerPort,
	workerPool backup.WorkerPoolPort,
) *BackupServiceAdapter {
	service := corebackup.NewService(config, storage, metrics, versioner, taskMgr, workerPool)
	return &BackupServiceAdapter{
		service: service,
	}
}

// Backup delegates to the core service
func (a *BackupServiceAdapter) Backup(ctx context.Context) error {
	return a.service.Backup(ctx)
}

// DryRun delegates to the core service
func (a *BackupServiceAdapter) DryRun(ctx context.Context) error {
	return a.service.DryRun(ctx)
}

// GetVersions delegates to the core service
func (a *BackupServiceAdapter) GetVersions() []backup.BackupVersion {
	return a.service.GetVersions()
}

// GetVersion delegates to the core service
func (a *BackupServiceAdapter) GetVersion(id string) (*backup.BackupVersion, error) {
	return a.service.GetVersion(id)
}

// GetLatestVersion delegates to the core service
func (a *BackupServiceAdapter) GetLatestVersion() (*backup.BackupVersion, error) {
	return a.service.GetLatestVersion()
}
