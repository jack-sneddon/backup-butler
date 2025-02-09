// internal/scan/compare.go
package scan

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"

	"github.com/jack-sneddon/backup-butler/internal/types"
)

func (s *Scanner) Compare(source, target string) ([]*FileComparison, error) {
	// Reset stats before scanning
	s.stats = make(map[string]*DirectoryStats)

	// Scan source
	s.log.Debugw("Starting source scan",
		"source", source,
		"excludePatterns", s.opts.ExcludePatterns)
	_, err := s.Scan(source)
	if err != nil {
		return nil, err
	}
	sourceStats := s.stats

	// Reset and scan target
	s.log.Debugw("Starting target scan",
		"target", target,
		"excludePatterns", s.opts.ExcludePatterns)
	s.stats = make(map[string]*DirectoryStats)
	_, err = s.Scan(target)
	if err != nil {
		return nil, err
	}
	targetStats := s.stats

	comparisons := make([]*FileComparison, 0)

	// Compare source files to target
	for _, dir := range sourceStats {
		for _, file := range dir.Files {
			relPath, err := filepath.Rel(source, file.Path)
			if err != nil {
				s.log.Debugw("Error getting relative path",
					"path", file.Path,
					"error", err)
				continue
			}

			s.log.Debugw("Checking file for comparison",
				"path", file.Path,
				"relPath", relPath,
				"excludePatterns", s.opts.ExcludePatterns)

			// Check if file should be excluded
			if matchesPattern(relPath, s.opts.ExcludePatterns) {
				s.log.Debugw("Excluding file from comparison",
					"relPath", relPath)
				continue
			}

			targetPath := filepath.Join(target, relPath)

			// Determine validation level based on path
			validationLevel := s.determineValidationLevel(relPath)

			comp := &FileComparison{
				Path:   relPath,
				Source: file,
				Level:  types.ValidationLevel(validationLevel),
			}

			if tf := findFile(targetStats, targetPath); tf != nil {
				comp.Target = tf

				// Do initial comparison at current level
				s.log.Debugw("Performing initial comparison",
					"path", file.Path,
					"level", comp.Level)

				comp.Status = s.compareFiles(file, tf, comp.Level)

				// If comparison shows differences and escalation is configured
				if comp.Status == StatusDiffer &&
					s.opts.ValidationConfig != nil &&
					s.opts.ValidationConfig.OnMismatch != "" {

					// Update level for escalated comparison
					oldLevel := comp.Level
					comp.Level = s.opts.ValidationConfig.OnMismatch
					s.log.Debugw("Escalating validation",
						"path", file.Path,
						"fromLevel", oldLevel,
						"toLevel", comp.Level)

					// Perform comparison at escalated level
					comp.Status = s.compareFiles(file, tf, comp.Level)
				}
			} else {
				comp.Status = StatusNew
			}
			comparisons = append(comparisons, comp)
		}
	}

	// Find target-only files
	for _, dir := range targetStats {
		for _, file := range dir.Files {
			relPath, err := filepath.Rel(target, file.Path)
			if err != nil {
				s.log.Debugw("Error getting relative path for target file",
					"path", file.Path,
					"error", err)
				continue
			}

			s.log.Debugw("Checking target file",
				"path", file.Path,
				"relPath", relPath,
				"excludePatterns", s.opts.ExcludePatterns)

			// Check if file should be excluded
			if matchesPattern(relPath, s.opts.ExcludePatterns) {
				s.log.Debugw("Excluding target file from comparison",
					"relPath", relPath)
				continue
			}

			if findFile(sourceStats, filepath.Join(source, relPath)) == nil {
				comparisons = append(comparisons, &FileComparison{
					Path:   relPath,
					Target: file,
					Status: StatusMissing,
				})
			}
		}
	}

	return comparisons, nil
}

func (s *Scanner) determineValidationLevel(path string) types.ValidationLevel {
	// Return the default validation level
	s.log.Debugw("Using default validation level",
		"path", path,
		"level", s.opts.DefaultLevel)
	return s.opts.DefaultLevel
}

func findFile(stats map[string]*DirectoryStats, path string) *FileInfo {
	dir := filepath.Dir(path)
	if dirStat, ok := stats[dir]; ok {
		for _, file := range dirStat.Files {
			if file.Path == path {
				return file
			}
		}
	}
	return nil
}

func (s *Scanner) compareFiles(src, tgt *FileInfo, level types.ValidationLevel) FileStatus {
	s.log.Debugw("Starting file comparison",
		"path", src.Path,
		"level", level)

	// First check metadata for all levels
	if src.Size != tgt.Size {
		s.log.Debugw("Size mismatch",
			"path", src.Path,
			"sourceSize", src.Size,
			"targetSize", tgt.Size)
		return StatusDiffer
	}

	// Compare modification times with tolerance
	const modTimeToleranceSeconds = 2
	if diff := abs(src.ModTime - tgt.ModTime); diff > modTimeToleranceSeconds {
		s.log.Debugw("Modification time mismatch",
			"path", src.Path,
			"sourceTime", src.ModTime,
			"targetTime", tgt.ModTime,
			"difference", diff)
		return StatusDiffer
	}

	// For Quick validation, metadata match is enough
	if level == types.Quick {
		s.log.Debugw("Quick validation metadata match",
			"path", src.Path)
		return StatusDiffer // Return StatusDiffer to trigger escalation
	}

	// For Standard and Deep validation, do hash comparison
	s.log.Debugw("Performing hash comparison",
		"path", src.Path,
		"level", level)

	srcHash, err := hashFile(src.Path)
	if err != nil {
		s.log.Debugw("Hash error", "path", src.Path, "error", err)
		return StatusError
	}
	tgtHash, err := hashFile(tgt.Path)
	if err != nil {
		s.log.Debugw("Hash error", "path", tgt.Path, "error", err)
		return StatusError
	}

	if srcHash != tgtHash {
		s.log.Debugw("Content mismatch",
			"path", src.Path,
			"level", level,
			"srcHash", srcHash[:8],
			"tgtHash", tgtHash[:8])
		return StatusDiffer
	}

	s.log.Debugw("File comparison successful",
		"path", src.Path,
		"level", level)
	return StatusMatch
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return string(h.Sum(nil)), nil
}
