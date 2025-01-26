// Package scan provides directory traversal and file analysis functionality for backup operations.
//
// The scanner is designed to optimize disk access patterns by:
// - Grouping files by directory to minimize disk head movement
// - Building directory statistics for efficient batch operations
// - Supporting resumable operations through stateful scanning
//
// Key Components:
// - FileInfo: Individual file metadata including path, size, and modification time
// - DirectoryStats: Aggregated directory information including total size and file count
// - Scanner: Main scanning engine that traverses directories and builds stats
//
// Usage:
//
//	scanner := scan.NewScanner()
//	err := scanner.Scan("/path/to/source")
//
//	// Access directory statistics
//	for dir, stats := range scanner.GetStats() {
//	    fmt.Printf("Dir: %s, Files: %d, Size: %d\n",
//	        dir, stats.FileCount, stats.TotalSize)
//	}
//
// Performance Considerations:
// - Groups files by directory to optimize for HDD access patterns
// - Maintains directory hierarchy for efficient batch operations
// - Supports incremental scanning for large directories
//
// The scanner is particularly optimized for HDDs by minimizing random access
// and grouping operations by directory to reduce disk head movement.
// internal/scan/scan.go
package scan

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/jack-sneddon/backup-butler/internal/logger"
	"go.uber.org/zap"
)

type FileInfo struct {
	Path    string
	Size    int64
	ModTime int64
	IsDir   bool
	Parent  string
}

type DirectoryStats struct {
	Path      string
	FileCount int
	TotalSize int64
	Files     []*FileInfo
}

type Progress struct {
	CurrentDir     string
	ScannedDirs    int
	ScannedFiles   int
	ProcessedBytes int64
	TotalBytes     int64
}

type Scanner struct {
	stats    map[string]*DirectoryStats
	log      *zap.SugaredLogger
	progress *Progress
	rootPath string
}

type FileStatus byte

const (
	StatusMatch   FileStatus = '=' // File identical
	StatusNew     FileStatus = '+' // Only in source
	StatusMissing FileStatus = '-' // Only in target
	StatusDiffer  FileStatus = '*' // Content differs
	StatusError   FileStatus = '!' // Error reading/comparing
)

type FileComparison struct {
	Path   string
	Status FileStatus
	Source *FileInfo
	Target *FileInfo
}

func NewScanner() *Scanner {
	return &Scanner{
		stats:    make(map[string]*DirectoryStats),
		log:      logger.Get(),
		progress: &Progress{},
	}
}

func (s *Scanner) Scan(root string) (*Progress, error) {
	s.rootPath = root

	// First pass - get total size
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			s.progress.TotalBytes += info.Size()
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Second pass - detailed scan
	return s.progress, filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		s.progress.CurrentDir = filepath.Dir(path)
		if info.IsDir() {
			s.progress.ScannedDirs++
		} else {
			s.progress.ScannedFiles++
			s.progress.ProcessedBytes += info.Size()
		}

		parent := filepath.Dir(path)
		if _, exists := s.stats[parent]; !exists {
			s.stats[parent] = &DirectoryStats{Path: parent}
		}

		dirStats := s.stats[parent]
		dirStats.FileCount++
		if !info.IsDir() {
			dirStats.TotalSize += info.Size()
			dirStats.Files = append(dirStats.Files, &FileInfo{
				Path:    path,
				Size:    info.Size(),
				ModTime: info.ModTime().Unix(),
				IsDir:   info.IsDir(),
				Parent:  parent,
			})
		}

		return nil
	})
}

func (s *Scanner) GetDirectoryStats() []*DirectoryStats {
	stats := make([]*DirectoryStats, 0, len(s.stats))
	for _, stat := range s.stats {
		stats = append(stats, stat)
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Path < stats[j].Path
	})
	return stats
}
