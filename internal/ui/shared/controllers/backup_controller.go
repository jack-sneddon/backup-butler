// internal/ui/shared/controllers/backup_controller.go
package controllers

import (
	"context"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
	"github.com/jack-sneddon/backup-butler/internal/ui/shared/events"
	"github.com/jack-sneddon/backup-butler/internal/ui/shared/viewmodels"
)

type BackupController struct {
	service  backup.BackupService
	eventBus *events.EventBus
	progress viewmodels.BackupProgress
}

func NewBackupController(service backup.BackupService, bus *events.EventBus) *BackupController {
	return &BackupController{
		service:  service,
		eventBus: bus,
	}
}

func (c *BackupController) StartBackup(ctx context.Context, config viewmodels.ConfigViewModel) error {
	c.eventBus.Publish(events.Event{
		Type: events.BackupStarted,
		Payload: viewmodels.BackupOperation{
			IsRunning: true,
			IsDryRun:  false,
		},
	})

	go func() {
		err := c.service.Backup(ctx)
		if err != nil {
			events.PublishError(c.eventBus, err)
		}
	}()

	return nil
}

func (c *BackupController) StartDryRun(ctx context.Context, config viewmodels.ConfigViewModel) error {
	c.eventBus.Publish(events.Event{
		Type: events.BackupStarted,
		Payload: viewmodels.BackupOperation{
			IsRunning: true,
			IsDryRun:  true,
		},
	})

	go func() {
		err := c.service.DryRun(ctx)
		if err != nil {
			events.PublishError(c.eventBus, err)
		}
	}()

	return nil
}

func (c *BackupController) GetVersions() []viewmodels.VersionViewModel {
	versions := c.service.GetVersions()
	result := make([]viewmodels.VersionViewModel, len(versions))

	for i, v := range versions {
		result[i] = viewmodels.VersionViewModel{
			ID:           v.ID,
			TimeStamp:    v.Timestamp,
			Duration:     v.Duration,
			FileCount:    v.Stats.TotalFiles,
			Size:         v.Size,
			Status:       v.Status,
			FilesCopied:  v.Stats.FilesBackedUp,
			FilesSkipped: v.Stats.FilesSkipped,
			FilesFailed:  v.Stats.FilesFailed,
		}
	}

	return result
}
