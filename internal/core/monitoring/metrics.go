// internal/core/monitoring/metrics.go
package monitoring

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

const (
	progressInterval = 100 * time.Millisecond
	progressBarWidth = 30
	megabyte         = 1024 * 1024
)

// Metrics handles tracking and reporting of backup progress
type Metrics struct {
	mu            sync.RWMutex
	totalFiles    int
	filesComplete int
	bytesComplete int64
	filesSkipped  int
	filesFailed   int
	startTime     time.Time
	quiet         bool
	cancelFunc    context.CancelFunc
}

func NewMetrics(quiet bool) *Metrics {
	return &Metrics{
		startTime: time.Now(),
		quiet:     quiet,
	}
}

func (m *Metrics) SetTotalFiles(total int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalFiles = total
}

func (m *Metrics) StartTracking(ctx context.Context) {
	m.mu.Lock()
	m.resetMetrics()
	trackingCtx, cancel := context.WithCancel(ctx)
	m.cancelFunc = cancel
	m.mu.Unlock()

	if !m.quiet {
		go m.trackProgress(trackingCtx)
	}
}

func (m *Metrics) StopTracking() {
	if m.cancelFunc != nil {
		m.cancelFunc()
	}
}

func (m *Metrics) resetMetrics() {
	m.filesComplete = 0
	m.filesSkipped = 0
	m.filesFailed = 0
	m.bytesComplete = 0
	m.startTime = time.Now()
}

func (m *Metrics) trackProgress(ctx context.Context) {
	ticker := time.NewTicker(progressInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.DisplayProgress()
		}
	}
}

func (m *Metrics) IncrementCompleted(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.filesComplete++
	m.bytesComplete += bytes
}

func (m *Metrics) IncrementSkipped(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.filesSkipped++
	m.bytesComplete += bytes
}

func (m *Metrics) IncrementFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.filesFailed++
}

func (m *Metrics) GetStats() backup.BackupStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return backup.BackupStats{
		TotalFiles:       m.totalFiles,
		FilesBackedUp:    m.filesComplete,
		FilesSkipped:     m.filesSkipped,
		FilesFailed:      m.filesFailed,
		TotalBytes:       m.bytesComplete,
		BytesTransferred: m.bytesComplete,
	}
}

func (m *Metrics) DisplayProgress() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.totalFiles == 0 {
		return
	}

	total := m.filesComplete + m.filesSkipped
	percentComplete := (float64(total) / float64(m.totalFiles)) * 100.0
	if total == m.totalFiles {
		percentComplete = 100.0
	}

	// Create progress bar
	completed := int((percentComplete / 100.0) * float64(progressBarWidth))
	completed = clamp(completed, 0, progressBarWidth)
	bar := createProgressBar(completed)

	// Calculate transfer speed
	speed := m.calculateTransferSpeed()

	// Display progress
	fmt.Print("\x1b[s")     // Save cursor position
	fmt.Print("\x1b[1000D") // Move cursor far left
	fmt.Print("\x1b[K")     // Clear line
	fmt.Printf("[%s] %5.1f%% | %3d copied, %3d skipped of %3d files | %6.2f MB | %6.2f MB/s",
		bar,
		percentComplete,
		m.filesComplete,
		m.filesSkipped,
		m.totalFiles,
		float64(m.bytesComplete)/megabyte,
		speed)
	fmt.Print("\x1b[u") // Restore cursor position
}

func (m *Metrics) DisplayFinalSummary() {
	if m.quiet {
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	duration := time.Since(m.startTime)
	fmt.Printf("\n\nBackup completed in %v\n", duration)
	fmt.Printf("Files processed: %d, Files skipped: %d, Failed: %d, Total size: %.2f MB\n",
		m.filesComplete,
		m.filesSkipped,
		m.filesFailed,
		float64(m.bytesComplete)/megabyte)
}

func (m *Metrics) calculateTransferSpeed() float64 {
	duration := time.Since(m.startTime).Seconds()
	if duration > 0 {
		return float64(m.bytesComplete) / duration / megabyte
	}
	return 0
}

func createProgressBar(completed int) string {
	return strings.Repeat("█", completed) + strings.Repeat("░", progressBarWidth-completed)
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
