// internal/core/backup/stats.go
package backup

import (
	"fmt"
	"path/filepath"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

// DryRunStats holds statistics for dry run execution
type DryRunStats struct {
	FilesToCopy  int
	TotalBytes   int64
	SkippedFiles int
	SkippedBytes int64
	TotalFiles   int
}

// Print formats and displays the dry run statistics
func (s *DryRunStats) Print() {
	fmt.Printf("\nDry Run Summary:\n")
	fmt.Printf("Files to copy:  %d (%.2f MB)\n", s.FilesToCopy, float64(s.TotalBytes)/1024/1024)
	fmt.Printf("Files to skip:  %d (%.2f MB)\n", s.SkippedFiles, float64(s.SkippedBytes)/1024/1024)
	fmt.Printf("Total files:    %d\n", s.TotalFiles)
	fmt.Printf("Total size:     %.2f MB\n", float64(s.TotalBytes+s.SkippedBytes)/1024/1024)
}

// UpdateForFile updates stats based on whether a file will be copied or skipped
func (s *DryRunStats) UpdateForFile(task backup.BackupTask, willCopy bool) {
	if willCopy {
		s.FilesToCopy++
		s.TotalBytes += task.Size
	} else {
		s.SkippedFiles++
		s.SkippedBytes += task.Size
	}
}

// LogFileStatus logs the status of an individual file during dry run
func (s *DryRunStats) LogFileStatus(task backup.BackupTask, willCopy bool, quiet bool) {
	if quiet {
		return
	}

	if willCopy {
		fmt.Printf("Would copy: %s (%.2f MB)\n",
			filepath.Base(task.Source),
			float64(task.Size)/1024/1024)
	} else {
		fmt.Printf("Would skip: %s (%.2f MB)\n",
			filepath.Base(task.Source),
			float64(task.Size)/1024/1024)
	}
}
