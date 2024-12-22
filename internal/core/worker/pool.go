// internal/core/worker/pool.go
package worker

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

const (
	minWorkers       = 1
	maxBackoffJitter = time.Second
)

// TaskExecutor defines the interface for task execution
type TaskExecutor interface {
	ExecuteTask(backup.BackupTask) error
	ShouldSkipFile(backup.BackupTask) (bool, error)
}

// Pool manages a pool of workers for concurrent task execution
type Pool struct {
	workers       int
	taskExecutor  TaskExecutor
	retryAttempts int
	retryDelay    time.Duration
}

// NewPool creates a new worker pool
func NewPool(workers int, executor TaskExecutor, retryAttempts int, retryDelay time.Duration) *Pool {
	if workers < minWorkers {
		workers = minWorkers
	}

	return &Pool{
		workers:       workers,
		taskExecutor:  executor,
		retryAttempts: retryAttempts,
		retryDelay:    retryDelay,
	}
}

// Execute processes tasks using the worker pool
func (p *Pool) Execute(ctx context.Context, tasks []backup.BackupTask) <-chan backup.TaskResult {
	resultCh := make(chan backup.TaskResult, len(tasks))
	taskCh := make(chan backup.BackupTask, p.workers)

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go p.worker(ctx, &wg, taskCh, resultCh)
	}

	// Feed tasks to workers
	go p.feedTasks(ctx, tasks, taskCh)

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	return resultCh
}

func (p *Pool) worker(ctx context.Context, wg *sync.WaitGroup, taskCh <-chan backup.BackupTask, resultCh chan<- backup.TaskResult) {
	defer wg.Done()

	for task := range taskCh {
		select {
		case <-ctx.Done():
			return
		default:
			result := p.processTask(ctx, task)
			select {
			case resultCh <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (p *Pool) feedTasks(ctx context.Context, tasks []backup.BackupTask, taskCh chan<- backup.BackupTask) {
	defer close(taskCh)

	for _, task := range tasks {
		select {
		case taskCh <- task:
		case <-ctx.Done():
			return
		}
	}
}

func (p *Pool) processTask(ctx context.Context, task backup.BackupTask) backup.TaskResult {
	// Check if task should be skipped
	shouldSkip, err := p.taskExecutor.ShouldSkipFile(task)
	if err != nil {
		return backup.TaskResult{
			Task:   task,
			Status: "failed",
			Error:  fmt.Errorf("skip check failed: %w", err),
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
	if err := p.executeWithRetry(ctx, task); err != nil {
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

func (p *Pool) executeWithRetry(ctx context.Context, task backup.BackupTask) error {
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
					// Calculate backoff with jitter
					backoff := p.retryDelay * time.Duration(attempt*attempt)
					jitter := time.Duration(rand.Int63n(int64(maxBackoffJitter)))

					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(backoff + jitter):
						// Continue to next attempt
					}
				}
			}
		}
	}

	return fmt.Errorf("task failed after %d attempts: %w", p.retryAttempts, lastErr)
}
