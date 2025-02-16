// internal/progress/tracker.go
package progress

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/logger"
	"go.uber.org/zap"
)

type tracker struct {
	progress *Progress
	display  *display
	mu       sync.Mutex
	log      *zap.SugaredLogger
}

func NewTracker() *tracker {
	p := &Progress{
		StartTime: time.Now(),
		Phase:     "initializing",
	}

	return &tracker{
		progress: p,
		display:  NewDisplay(p),
		log:      logger.Get(),
	}
}

func (t *tracker) Start() error {
	t.log.Debugw("Starting progress tracker")
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

	if t.progress.Current != nil {
		t.progress.Current.Processed += bytes
		t.progress.Current.Done++
		t.progress.BytesDone += bytes
		t.progress.Processed++

		t.log.Debugw("Progress updated",
			"directory", t.progress.Current.Path,
			"processed", t.progress.Processed,
			"total", t.progress.TotalFiles,
			"bytes", t.progress.BytesDone)

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

	// Make sure we're tracking total files across all directories
	t.progress.TotalFiles += fileCount  // Add to total
	t.progress.TotalBytes += totalBytes // Add to total bytes

	t.progress.Current = &DirectoryProgress{
		Path:      path,
		Total:     totalBytes,
		Files:     fileCount,
		StartTime: time.Now(),
	}

	t.log.Debugw("Starting directory",
		"path", path,
		"totalBytes", totalBytes,
		"fileCount", fileCount,
		"overallTotal", t.progress.TotalFiles)

	return nil
}
