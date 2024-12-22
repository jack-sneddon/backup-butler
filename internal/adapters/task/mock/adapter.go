package mock

import (
	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type MockTaskManager struct {
	CreateTasksFunc    func(config *backup.BackupConfig) ([]backup.BackupTask, int, error)
	ShouldSkipFileFunc func(task backup.BackupTask) (bool, error)
	ExecuteTaskFunc    func(task backup.BackupTask) error
	ValidateTaskFunc   func(task backup.BackupTask) error
}

func NewMockTaskManager() *MockTaskManager {
	return &MockTaskManager{
		CreateTasksFunc: func(config *backup.BackupConfig) ([]backup.BackupTask, int, error) {
			return []backup.BackupTask{}, 0, nil
		},
		ShouldSkipFileFunc: func(task backup.BackupTask) (bool, error) {
			return false, nil
		},
		ExecuteTaskFunc: func(task backup.BackupTask) error {
			return nil
		},
		ValidateTaskFunc: func(task backup.BackupTask) error {
			return nil
		},
	}
}

func (m *MockTaskManager) CreateTasks(config *backup.BackupConfig) ([]backup.BackupTask, int, error) {
	return m.CreateTasksFunc(config)
}

func (m *MockTaskManager) ShouldSkipFile(task backup.BackupTask) (bool, error) {
	return m.ShouldSkipFileFunc(task)
}

func (m *MockTaskManager) ExecuteTask(task backup.BackupTask) error {
	return m.ExecuteTaskFunc(task)
}

func (m *MockTaskManager) ValidateTask(task backup.BackupTask) error {
	return m.ValidateTaskFunc(task)
}
