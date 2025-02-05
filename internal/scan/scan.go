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
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/jack-sneddon/backup-butler/internal/logger"
	"go.uber.org/zap"
)

type Scanner struct {
	stats    map[string]*DirectoryStats
	log      *zap.SugaredLogger
	progress *Progress
	rootPath string
	opts     *ScannerOptions
	mu       sync.Mutex // Protects stats map
}

func NewScanner(options *ScannerOptions) *Scanner {
	if options == nil {
		options = &ScannerOptions{
			MaxDepth:     -1,
			BufferSize:   32768,
			DefaultLevel: "standard", // Default if not specified
		}
	}

	// Set validation defaults if not provided
	if options.ValidationConfig == nil {
		options.ValidationConfig = &ValidationConfig{
			DefaultLevel: "standard",
		}
	}

	return &Scanner{
		stats:    make(map[string]*DirectoryStats),
		log:      logger.Get(),
		progress: &Progress{Phase: "initializing"},
		opts:     options,
	}
}

// GetProgress returns the current progress information
func (s *Scanner) GetProgress() *Progress {
	return s.progress
}

func (s *Scanner) Scan(root string) (*Progress, error) {
	s.rootPath = root
	s.progress.Phase = "counting"

	// Reset all counters
	s.progress.TotalFiles = 0
	s.progress.TotalBytes = 0
	s.progress.ScannedFiles = 0
	s.progress.ScannedDirs = 0
	s.progress.ProcessedBytes = 0
	s.progress.ExcludedFiles = 0
	s.progress.ExcludedDirs = 0

	// First pass - count total files and size
	if err := s.countFiles(root); err != nil {
		return nil, err
	}

	s.log.Debugw("Count complete",
		"totalFiles", s.progress.TotalFiles,
		"totalBytes", s.progress.TotalBytes,
		"excludedFiles", s.progress.ExcludedFiles)

	// Second pass - detailed scan
	if err := s.scanFiles(root, 0); err != nil {
		return nil, err
	}

	return s.progress, nil
}

func (s *Scanner) countFiles(root string) error {
	// Convert root to absolute path
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	s.log.Debugw("Starting file count",
		"root", absRoot,
		"excludePatterns", s.opts.ExcludePatterns,
		"includeFolders", s.opts.IncludeFolders)

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.progress.AddError(NewScanError(path, "access", err))
			return nil // Continue despite errors
		}

		/*
			s.log.Debugw("Processing path",
				"path", path,
				"isDir", info.IsDir(),
				"root", absRoot)
		*/

		if info.IsDir() {
			// Skip directory pattern checks for root
			if path != absRoot {
				if !shouldIncludeFolder(path, s.opts.IncludeFolders) {
					s.log.Debugw("Excluding directory by folder list", "path", path)
					s.progress.ExcludedDirs++
					return filepath.SkipDir
				}
				// Get relative path for directory
				relPath, err := filepath.Rel(absRoot, path)
				if err != nil {
					s.progress.AddError(NewScanError(path, "rel_path", err))
					return nil
				}
				if matchesPattern(relPath, s.opts.ExcludePatterns) {
					s.log.Debugw("Excluding directory by pattern",
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
		if len(s.opts.ExcludePatterns) > 0 {
			relPath, err := filepath.Rel(absRoot, path)
			if err != nil {
				s.progress.AddError(NewScanError(path, "rel_path", err))
				return nil
			}
			/*
				s.log.Debugw("Checking file against patterns",
					"relPath", relPath,
					"patterns", s.opts.ExcludePatterns)
			*/

			if shouldExclude := matchesPattern(relPath, s.opts.ExcludePatterns); shouldExclude {
				/*
					s.log.Debugw("Excluding file by pattern",
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
			s.log.Debugw("Including file",
				"path", path,
				"size", info.Size(),
				"totalFiles", s.progress.TotalFiles,
				"totalBytes", s.progress.TotalBytes)
		*/

		return nil
	})
}

func (s *Scanner) scanFiles(root string, depth int) error {
	// Convert root to absolute path
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	if s.opts.MaxDepth >= 0 && depth > s.opts.MaxDepth {
		return nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		s.progress.AddError(NewScanError(root, "read_dir", err))
		return nil
	}

	s.progress.CurrentDir = root

	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		info, err := entry.Info()
		if err != nil {
			s.progress.AddError(NewScanError(path, "stat", err))
			continue
		}

		if info.IsDir() {
			// Skip directory pattern checks for root
			if path != absRoot {
				if !shouldIncludeFolder(path, s.opts.IncludeFolders) {
					continue
				}
				// Get relative path for directory
				relPath, err := filepath.Rel(absRoot, path)
				if err != nil {
					s.progress.AddError(NewScanError(path, "rel_path", err))
					continue
				}
				if matchesPattern(relPath, s.opts.ExcludePatterns) {
					continue
				}
			}
			if err := s.scanFiles(path, depth+1); err != nil {
				s.progress.AddError(err)
			}
			continue
		}

		// Get relative path for file
		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			s.progress.AddError(NewScanError(path, "rel_path", err))
			continue
		}

		if matchesPattern(relPath, s.opts.ExcludePatterns) {
			continue
		}

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

		s.progress.ScannedFiles++
		s.progress.ProcessedBytes += info.Size()
	}

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
