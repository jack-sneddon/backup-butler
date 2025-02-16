// internal/progress/display.go
package progress

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/logger"
)

type display struct {
	progress *Progress
}

func NewDisplay(p *Progress) *display {
	return &display{
		progress: p,
	}
}

func (d *display) Start() {
	logger.Debug("Display starting")
}

func (d *display) Stop() {
	logger.Debug("Display stopping")
}

func (d *display) render() {
	logger.Debug("Rendering display",
		"current", d.progress.Current,
		"totalFiles", d.progress.TotalFiles,
		"processed", d.progress.Processed)

	if d.progress.Current == nil {
		return
	}

	// Current directory - match test expectation
	fmt.Printf("\nProcessing: %s\n", d.progress.Current.Path)

	// Progress bar
	pct := 0.0
	if d.progress.Current.Total > 0 {
		pct = float64(d.progress.Current.Processed) / float64(d.progress.Current.Total) * 100
	}
	fmt.Printf("[%s] %.1f%% (%d/%d files)\n",
		renderBar(pct, 40),
		pct,
		d.progress.Current.Done,
		d.progress.Current.Files)

	fmt.Printf("Currently Processing:\n")
	fmt.Printf("  %s (%.1f MB)\n",
		filepath.Base(d.progress.Current.Path),
		float64(d.progress.Current.Total)/(1024*1024))

	// Statistics
	fmt.Printf("\nStatistics:\n")
	fmt.Printf("├── Processed: %d files (%.1f GB)\n",
		d.progress.Processed,
		float64(d.progress.BytesDone)/(1024*1024*1024))
	fmt.Printf("├── Remaining: %d files (%.1f GB)\n",
		d.progress.TotalFiles-d.progress.Processed,
		float64(d.progress.TotalBytes-d.progress.BytesDone)/(1024*1024*1024))
	fmt.Printf("└── Total Time: %s\n\n", // Changed from "Time" to "Total Time"
		time.Since(d.progress.StartTime).Round(time.Second))
}

func renderBar(percent float64, width int) string {
	filled := int(float64(width) * percent / 100)
	filled = min(filled, width) // Use min instead of if check

	full := strings.Repeat("=", filled)
	if filled < width {
		return full + ">" + strings.Repeat(" ", width-filled-1)
	}
	return full
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
