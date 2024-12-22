package pool

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type WorkerPool struct {
	workers       int
	taskExecutor  TaskExecutor
	retryAttempts int
	retryDelay    time.Duration
}

type TaskExecutor interface {
	ExecuteTask(backup.BackupTask) error
	ShouldSkipFile(backup.BackupTask) (bool, error)
}

func NewWorkerPool(
	workers int,
	executor TaskExecutor,
	retryAttempts int,
	retryDelay time.Duration,
	//logger backup.LoggerPort,
) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}
	return &WorkerPool{
		workers:       workers,
		taskExecutor:  executor,
		retryAttempts: retryAttempts,
		retryDelay:    retryDelay,
		//logger:        logger,
	}
}

func (p *WorkerPool) Execute(ctx context.Context, tasks []backup.BackupTask) <-chan backup.TaskResult {
	resultCh := make(chan backup.TaskResult, len(tasks))
	taskCh := make(chan backup.BackupTask, p.workers)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range taskCh {
				select {
				case <-ctx.Done():
					return
				default:
					result := p.processTask(ctx, task, workerID)
					select {
					case resultCh <- result:
					case <-ctx.Done():
						return
					}
				}
			}
		}(i)
	}

	// Feed tasks
	go func() {
		defer close(taskCh)
		for _, task := range tasks {
			select {
			case taskCh <- task:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for completion and close result channel
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	return resultCh
}

func (p *WorkerPool) processTask(ctx context.Context, task backup.BackupTask, workerID int) backup.TaskResult {
	// First check if we should skip
	shouldSkip, err := p.taskExecutor.ShouldSkipFile(task)
	if err != nil {
		//p.logger.Error("Worker %d: Failed to check if file should be skipped: %v", workerID, err)
		return backup.TaskResult{
			Task:   task,
			Status: "failed",
			Error:  err,
		}
	}

	if shouldSkip {
		return backup.TaskResult{
			Task:   task,
			Status: "skipped",
			Bytes:  task.Size,
		}
	}

	// Execute task with retry
	err = p.executeWithRetry(ctx, task)
	if err != nil {
		//p.logger.Error("Worker %d: Failed to execute task: %v", workerID, err)
		return backup.TaskResult{
			Task:   task,
			Status: "failed",
			Error:  err,
		}
	}

	return backup.TaskResult{
		Task:   task,
		Status: "completed",
		Bytes:  task.Size,
	}
}

func (p *WorkerPool) executeWithRetry(ctx context.Context, task backup.BackupTask) error {
	var lastErr error

	for attempt := 1; attempt <= p.retryAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := p.taskExecutor.ExecuteTask(task); err == nil {
				return nil
			} else {
				lastErr = err
				if attempt < p.retryAttempts {
					backoff := p.retryDelay * time.Duration(attempt*attempt)
					jitter := time.Duration(rand.Int63n(int64(time.Second)))
					time.Sleep(backoff + jitter)
				}
			}
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", p.retryAttempts, lastErr)
}
