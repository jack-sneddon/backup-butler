package collector

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type MetricsCollector struct {
	mu            sync.RWMutex
	totalFiles    int
	filesComplete int
	bytesComplete int64
	filesSkipped  int
	filesFailed   int
	startTime     time.Time
	quiet         bool
}

func NewMetricsCollector(totalFiles int, quiet bool) *MetricsCollector {
	return &MetricsCollector{
		totalFiles: totalFiles,
		startTime:  time.Now(),
		quiet:      quiet,
	}
}

func (m *MetricsCollector) StartTracking(ctx context.Context) {
	// Reset metrics
	m.mu.Lock()
	m.filesComplete = 0
	m.filesSkipped = 0
	m.filesFailed = 0
	m.bytesComplete = 0
	m.startTime = time.Now()
	m.mu.Unlock()

	// Start progress display ticker
	if !m.quiet {
		ticker := time.NewTicker(100 * time.Millisecond)
		go func() {
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					m.DisplayProgress()
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

func (m *MetricsCollector) IncrementCompleted(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.filesComplete++
	m.bytesComplete += bytes
}

func (m *MetricsCollector) IncrementSkipped(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.filesSkipped++
	m.bytesComplete += bytes
}

func (m *MetricsCollector) IncrementFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.filesFailed++
}

func (m *MetricsCollector) GetStats() backup.BackupStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Verify total matches
	total := m.filesComplete + m.filesSkipped + m.filesFailed
	if total != m.totalFiles {
		log.Printf("Warning: Total files mismatch - Expected: %d, Got: %d (Completed: %d, Skipped: %d, Failed: %d)",
			m.totalFiles, total, m.filesComplete, m.filesSkipped, m.filesFailed)
	}

	return backup.BackupStats{
		TotalFiles:       m.totalFiles,
		FilesBackedUp:    m.filesComplete,
		FilesSkipped:     m.filesSkipped,
		FilesFailed:      m.filesFailed,
		TotalBytes:       m.bytesComplete,
		BytesTransferred: m.bytesComplete,
	}
}

func (m *MetricsCollector) DisplayProgress() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := m.filesComplete + m.filesSkipped
	if m.totalFiles == 0 {
		return // avoid division by zero
	}

	percentComplete := (float64(total) / float64(m.totalFiles)) * 100.0
	if total == m.totalFiles {
		percentComplete = 100.0
	}

	const barWidth = 30
	completed := int((percentComplete / 100.0) * float64(barWidth))
	if completed < 0 {
		completed = 0
	}
	if completed > barWidth {
		completed = barWidth
	}

	bar := strings.Repeat("█", completed) + strings.Repeat("░", barWidth-completed)

	duration := time.Since(m.startTime).Seconds()
	var speed float64
	if duration > 0 {
		speed = float64(m.bytesComplete) / duration / 1024 / 1024 // MB/s
	}

	fmt.Print("\x1b[s")     // Save cursor position
	fmt.Print("\x1b[1000D") // Move cursor far left
	fmt.Print("\x1b[K")     // Clear line
	fmt.Printf("[%s] %5.1f%% | %3d copied, %3d skipped of %3d files | %6.2f MB | %6.2f MB/s",
		bar,
		percentComplete,
		m.filesComplete,
		m.filesSkipped,
		m.totalFiles,
		float64(m.bytesComplete)/1024/1024,
		speed)
	fmt.Print("\x1b[u") // Restore cursor position
}

func (m *MetricsCollector) DisplayFinalSummary() {
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
		float64(m.bytesComplete)/1024/1024)
}

func (m *MetricsCollector) SetTotalFiles(total int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalFiles = total
}
