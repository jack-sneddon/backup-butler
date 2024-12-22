// internal/adapters/worker/pool/adapter.go
package pool

import (
	"context"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/core/worker"
	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type WorkerPoolAdapter struct {
	pool *worker.Pool
}

func NewWorkerPool(
	workers int,
	executor worker.TaskExecutor,
	retryAttempts int,
	retryDelay time.Duration,
) *WorkerPoolAdapter {
	return &WorkerPoolAdapter{
		pool: worker.NewPool(workers, executor, retryAttempts, retryDelay),
	}
}

func (a *WorkerPoolAdapter) Execute(ctx context.Context, tasks []backup.BackupTask) <-chan backup.TaskResult {
	return a.pool.Execute(ctx, tasks)
}
