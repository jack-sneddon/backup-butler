// Package scan provides directory traversal, file analysis, and validation functionality for backup operations.
//
// The scanner is designed with three validation levels:
// - Quick: File metadata comparison (size, modification time)
// - Standard: Metadata plus partial content hash (configurable buffer)
// - Deep: Full file content validation
//
// The scanner optimizes disk access patterns by:
// - Grouping files by directory to minimize disk head movement
// - Building directory statistics for efficient batch operations
// - Supporting resumable operations through stateful scanning
//
// Key Components:
// - FileInfo: Individual file metadata including path, size, modification time
// - DirectoryStats: Aggregated directory information including total size and file count
// - Scanner: Main scanning engine that traverses directories and performs validation
// - ValidatorStrategy: Interface for different validation levels
//
// Usage:
//
//	opts := &ScannerOptions{
//	    ExcludePatterns: []string{"*.tmp"},
//	    IncludeFolders:  []string{"photos", "documents"},
//	    MaxDepth:        -1,
//	    BufferSize:      32768,
//	}
//	scanner := scan.NewScanner(opts)
//	progress, err := scanner.Scan("/path/to/source")
//
// Performance Considerations:
// - Groups files by directory to optimize for HDD access patterns
// - Maintains directory hierarchy for efficient batch operations
// - Uses configurable validation levels to balance speed vs confidence
// - Supports incremental scanning for large directories
//
// The scanner is particularly optimized for HDDs by:
// - Minimizing random access through directory grouping
// - Using configurable buffer sizes for partial hash validation
// - Supporting staged validation from quick to deep
package scan

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/logger"
	"github.com/jack-sneddon/backup-butler/internal/types"
)

type Scanner struct {
	stats    map[string]*DirectoryStats
	progress *Progress
	rootPath string
	opts     *ScannerOptions
	mu       sync.Mutex // Protects stats map
}

// ScannerOptions defines configuration options for Scanner
type ScannerOptions struct {
	ExcludePatterns  []string
	IncludeFolders   []string
	MaxDepth         int
	BufferSize       int
	Level            types.ValidationLevel
	ValidationConfig *ValidationConfig
}

func NewScanner(options *ScannerOptions) *Scanner {
	if options == nil {
		options = &ScannerOptions{
			MaxDepth:   -1,
			BufferSize: 32768,
			Level:      types.Standard,
		}
	}

	if options.ValidationConfig == nil {
		options.ValidationConfig = &ValidationConfig{
			Level: types.Standard,
		}
	}

	return &Scanner{
		stats: make(map[string]*DirectoryStats),
		progress: &Progress{
			Phase:     "initializing",
			StartTime: time.Now(),
		},
		opts: options,
	}
}

// GetProgress returns the current progress information
func (s *Scanner) GetProgress() *Progress {
	return s.progress
}

func (s *Scanner) Scan(root string) (*Progress, error) {
	scanLogger := logger.WithGroup("scanner")

	scanLogger.Info("Starting scan operation",
		"root", root,
		"level", s.opts.Level)

	s.rootPath = root
	s.progress.Phase = "counting"

	// Reset all counters
	scanLogger.Debug("Resetting scan counters")

	s.progress.TotalFiles = 0
	s.progress.TotalBytes = 0
	s.progress.ScannedFiles = 0
	s.progress.ScannedDirs = 0
	s.progress.ProcessedBytes = 0
	s.progress.ExcludedFiles = 0
	s.progress.ExcludedDirs = 0

	// First pass - count total files and size
	scanLogger.Info("Starting file count phase")
	if err := s.countFiles(root); err != nil {
		return nil, err
	}

	scanLogger.Info("Count complete",
		"totalFiles", s.progress.TotalFiles,
		"totalBytes", s.progress.TotalBytes,
		"excludedFiles", s.progress.ExcludedFiles)

	// Second pass - detailed scan
	scanLogger.Info("Starting detailed scan phase")
	if err := s.scanFiles(root, 0); err != nil {
		return nil, err
	}

	return s.progress, nil
}

func (s *Scanner) countFiles(root string) error {
	scanLogger := logger.WithGroup("scanner").With(
		"root", root,
		"level", s.opts.Level,
		"maxDepth", s.opts.MaxDepth,
	)
	scanLogger.Info("Starting countFiles operation")

	// Convert root to absolute path
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	logger.Debug("Starting file count",
		"root", absRoot,
		"excludePatterns", s.opts.ExcludePatterns,
		"includeFolders", s.opts.IncludeFolders)

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		pathLogger := scanLogger.With("path", path)
		if err != nil {
			pathLogger.Error("Access error", "error", err)
			s.progress.AddError(NewScanError(path, "access", err))
			return nil // Continue despite errors
		}

		if info.IsDir() {
			// Skip directory pattern checks for root
			if path != absRoot {
				if !shouldIncludeFolder(path, s.opts.IncludeFolders) {
					pathLogger.Debug("Excluding directory by folder list")
					s.progress.ExcludedDirs++
					return filepath.SkipDir
				}
				pathLogger.Debug("Processing directory")
				// Get relative path for directory
				relPath, err := filepath.Rel(absRoot, path)
				if err != nil {
					s.progress.AddError(NewScanError(path, "rel_path", err))
					return nil
				}
				if matchesPattern(relPath, s.opts.ExcludePatterns) {
					logger.Debug("Excluding directory by pattern",
						"path", path,
						"relPath", relPath)
					s.progress.ExcludedDirs++
					return filepath.SkipDir
				}
			}
			s.progress.ScannedDirs++
			return nil
		}

		// Handle files
		pathLogger.Debug("Processing file",
			"size", info.Size(),
			"modTime", info.ModTime(),
		)
		if len(s.opts.ExcludePatterns) > 0 {
			relPath, err := filepath.Rel(absRoot, path)
			if err != nil {
				s.progress.AddError(NewScanError(path, "rel_path", err))
				return nil
			}
			/*
				s.logger.Debug("Checking file against patterns",
					"relPath", relPath,
					"patterns", s.opts.ExcludePatterns)
			*/

			if shouldExclude := matchesPattern(relPath, s.opts.ExcludePatterns); shouldExclude {
				/*
					s.logger.Debug("Excluding file by pattern",
						"path", path,
						"relPath", relPath,
						"patterns", s.opts.ExcludePatterns)
				*/
				s.progress.ExcludedFiles++
				return nil
			}
		}

		// Include the file in totals
		s.progress.TotalFiles++
		s.progress.TotalBytes += info.Size()
		/*
			s.logger.Debug("Including file",
				"path", path,
				"size", info.Size(),
				"totalFiles", s.progress.TotalFiles,
				"totalBytes", s.progress.TotalBytes)
		*/

		return nil
	})
}

func (s *Scanner) scanFiles(root string, depth int) error {
	scanLogger := logger.WithGroup("scanner").With("path", root)

	// Convert root to absolute path
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	// Skip if max depth exceeded
	if s.opts.MaxDepth >= 0 && depth > s.opts.MaxDepth {
		scanLogger.Debug("Max depth exceeded", "depth", depth)
		return nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		scanLogger.Error("Failed to read directory", "error", err)
		s.progress.AddError(NewScanError(root, "read_dir", err))
		return nil
	}

	s.progress.CurrentDir = root
	s.progress.CurrentDirStart = time.Now()

	// Calculate directory totals first
	var dirTotal int64
	var fileCount int

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				dirTotal += info.Size()
				fileCount++
			}
		}
	}

	// Directory start - DEBUG level
	scanLogger.Debug("Processing directory",
		"entries", len(entries))

	s.progress.CurrentDirTotal = dirTotal
	s.progress.CurrentDirFiles = fileCount
	s.progress.CurrentDirDone = 0
	s.progress.CurrentDirBytes = 0

	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		info, err := entry.Info()
		if err != nil {
			scanLogger.Error("Failed to get file info",
				"file", entry.Name(),
				"error", err)
			s.progress.AddError(NewScanError(path, "stat", err))
			continue
		}

		if info.IsDir() {
			// Skip directory pattern checks for root
			if path != absRoot {
				if !shouldIncludeFolder(path, s.opts.IncludeFolders) {
					// Add logging but keep existing logic
					scanLogger.Debug("Excluding directory",
						"path", path,
						"reason", "folder list")
					continue
				}
				// Get relative path for directory
				relPath, err := filepath.Rel(absRoot, path)
				if err != nil {
					s.progress.AddError(NewScanError(path, "rel_path", err))
					continue
				}
				if matchesPattern(relPath, s.opts.ExcludePatterns) {
					scanLogger.Debug("Excluding directory",
						"path", path,
						"reason", "pattern match")
					continue
				}
			}
			if err := s.scanFiles(path, depth+1); err != nil {
				s.progress.AddError(err)
			}
			continue
		}

		// File processing with progress tracking
		// Get relative path for file
		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			s.progress.AddError(NewScanError(path, "rel_path", err))
			continue
		}

		if matchesPattern(relPath, s.opts.ExcludePatterns) {
			continue
		}

		s.progress.ScannedFiles++
		s.progress.ProcessedBytes += info.Size()

		// Update directory-level progress
		s.progress.CurrentDirDone++
		s.progress.CurrentDirBytes += info.Size()

		logger.Debug("File processed",
			"directory", root,
			"progress", fmt.Sprintf("%d/%d", s.progress.CurrentDirDone, s.progress.CurrentDirFiles),
			"bytes", fmt.Sprintf("%d/%d", s.progress.CurrentDirBytes, s.progress.CurrentDirTotal))

		// Process file
		s.mu.Lock()
		parent := filepath.Dir(path)
		if _, exists := s.stats[parent]; !exists {
			s.stats[parent] = &DirectoryStats{Path: parent}
		}

		dirStats := s.stats[parent]
		dirStats.FileCount++
		dirStats.TotalSize += info.Size()
		dirStats.Files = append(dirStats.Files, &FileInfo{
			Path:    path,
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
			IsDir:   info.IsDir(),
			Parent:  parent,
		})

		s.mu.Unlock()

	}

	// Directory complete - DEBUG level
	scanLogger.Debug("Directory processing complete",
		"scannedFiles", s.progress.ScannedFiles,
		"processedBytes", s.progress.ProcessedBytes)

	return nil
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
