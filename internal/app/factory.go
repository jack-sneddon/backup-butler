package app

import (
	"fmt"

	configadapter "github.com/jack-sneddon/backup-butler/internal/adapters/config/file"
	"github.com/jack-sneddon/backup-butler/internal/adapters/metrics/collector"
	backupservice "github.com/jack-sneddon/backup-butler/internal/adapters/service/backup"
	storageadapter "github.com/jack-sneddon/backup-butler/internal/adapters/storage/filesystem"
	"github.com/jack-sneddon/backup-butler/internal/adapters/task/manager"
	versionadapter "github.com/jack-sneddon/backup-butler/internal/adapters/version/file"
	pooladapter "github.com/jack-sneddon/backup-butler/internal/adapters/worker/pool"
	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

// Factory handles the creation and wiring of all components
type Factory struct {
	configPath string
}

func NewFactory(configPath string) *Factory {
	return &Factory{
		configPath: configPath,
	}
}

func (f *Factory) CreateBackupService() (backup.BackupService, error) {
	// 1. Load configuration
	configLoader := configadapter.NewFileConfigLoader()
	config, err := configLoader.Load(f.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// 2. Create storage adapter
	//storage := storageadapter.NewFilesystemAdapter(config.ChecksumAlgorithm)
	// Create storage adapter
	storage := storageadapter.NewFilesystemAdapter(
		config.ChecksumAlgorithm,
		config.BufferSize,
	)

	// 3. Create version manager
	versioner, err := versionadapter.NewFileVersionManager(config.TargetDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to create version manager: %w", err)
	}

	// 4. Create metrics collector
	metrics := collector.NewMetricsCollector(0, config.Options.Quiet)

	// 5. Create task manager
	taskMgr := manager.NewTaskManager(storage, metrics)
	taskMgr.SetConfig(config)

	// 6. Create worker pool
	workerPool := pooladapter.NewWorkerPool(
		config.Concurrency,
		taskMgr,
		config.RetryAttempts,
		config.RetryDelay,
	)

	// 7. Create backup service
	service := backupservice.NewBackupService(
		config,
		storage,
		metrics,
		versioner,
		taskMgr,
		workerPool,
	)

	return service, nil
}
