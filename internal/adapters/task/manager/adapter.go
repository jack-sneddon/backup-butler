// internal/adapters/task/manager/adapter.go
package manager

import (
	"github.com/jack-sneddon/backup-butler/internal/core/task"
	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type TaskManager struct {
	manager *task.Manager
}

func NewTaskManager(
	storage backup.StoragePort,
	metrics backup.MetricsPort,
) *TaskManager {
	return &TaskManager{
		manager: task.NewManager(storage, metrics),
	}
}

func (t *TaskManager) SetConfig(config *backup.BackupConfig) {
	t.manager.SetConfig(config)
}

func (t *TaskManager) CreateTasks(config *backup.BackupConfig) ([]backup.BackupTask, int, error) {
	return t.manager.CreateTasks(config)
}

func (t *TaskManager) ShouldSkipFile(task backup.BackupTask) (bool, error) {
	return t.manager.ShouldSkipFile(task)
}

func (t *TaskManager) ExecuteTask(task backup.BackupTask) error {
	return t.manager.ExecuteTask(task)
}

func (t *TaskManager) ValidateTask(task backup.BackupTask) error {
	return t.manager.ValidateTask(task)
}
