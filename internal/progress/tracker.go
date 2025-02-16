// internal/progress/tracker.go
package progress

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/logger"
)

type tracker struct {
	progress *Progress
	display  *display
	mu       sync.Mutex
}

func NewTracker() *tracker {
	p := &Progress{
		StartTime: time.Now(),
		Phase:     "initializing",
	}

	return &tracker{
		progress: p,
		display:  NewDisplay(p),
	}
}

func (t *tracker) Start() error {
	logger.Debug("Starting progress tracker")

	t.display.Start()
	return nil
}

func (t *tracker) Stop() error {
	t.display.Stop()
	return nil
}

func (t *tracker) UpdateProgress(bytes int64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	progLogger := logger.WithGroup("progress")

	// Guard against nil Current
	if t.progress.Current == nil {
		progLogger.Warn("Attempted to update progress with nil Current directory")
		return
	}

	// Update progress - DEBUG level
	progLogger.Debug("Updating progress",
		"directory", t.progress.Current.Path,
		"bytes", bytes)

	if t.progress.Current != nil {
		t.progress.Current.Processed += bytes
		t.progress.Current.Done++
		t.progress.BytesDone += bytes
		t.progress.Processed++

		// Progress summary - DEBUG level
		progLogger.Debug("Progress updated",
			"dirProgress", fmt.Sprintf("%d/%d",
				t.progress.Current.Done,
				t.progress.Current.Files),
			"totalProgress", fmt.Sprintf("%d/%d",
				t.progress.Processed,
				t.progress.TotalFiles))

		// Update display immediately
		t.display.render()
	}
}

func (t *tracker) GetProgress() *Progress {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.progress
}

func (t *tracker) FinishDirectory() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.progress.Current != nil {
		logger.WithGroup("progress").Info("Directory processing complete",
			"path", t.progress.Current.Path,
			"processedFiles", t.progress.Current.Done,
			"processedBytes", t.progress.Current.Processed)

		t.progress.Current = nil
	}
	return nil
}

func (t *tracker) ScanDirectory(path string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.progress.Phase = "scanning"

	// Walk the directory to count files
	var totalFiles int
	var totalBytes int64

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalFiles++
			totalBytes += info.Size()
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	t.progress.TotalFiles = totalFiles
	t.progress.TotalBytes = totalBytes
	t.progress.Phase = "processing"

	return nil
}

func (t *tracker) StartDirectory(path string, totalBytes int64, fileCount int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	progLogger := logger.WithGroup("progress")

	// Directory start - INFO level
	progLogger.Info("Starting directory processing",
		"path", path,
		"totalBytes", totalBytes,
		"fileCount", fileCount)

	// Make sure we're tracking total files across all directories
	t.progress.TotalFiles += fileCount  // Add to total
	t.progress.TotalBytes += totalBytes // Add to total bytes

	t.progress.Current = &DirectoryProgress{
		Path:      path,
		Total:     totalBytes,
		Files:     fileCount,
		StartTime: time.Now(),
	}

	logger.Debug("Starting directory",
		"path", path,
		"totalBytes", totalBytes,
		"fileCount", fileCount,
		"overallTotal", t.progress.TotalFiles)

	return nil
}
