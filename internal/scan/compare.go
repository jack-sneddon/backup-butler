// internal/scan/compare.go
package scan

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"

	"github.com/jack-sneddon/backup-butler/internal/logger"
	"github.com/jack-sneddon/backup-butler/internal/types"
)

func (s *Scanner) Compare(source, target string) ([]*FileComparison, error) {
	s.stats = make(map[string]*DirectoryStats)

	// Scan source
	_, err := s.Scan(source)
	if err != nil {
		return nil, err
	}
	sourceStats := s.stats

	// Reset and scan target
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
				logger.Debug("Error getting relative path",
					"path", file.Path,
					"error", err)
				continue
			}

			if matchesPattern(relPath, s.opts.ExcludePatterns) {
				logger.Debug("Excluding file from comparison",
					"relPath", relPath)
				continue
			}

			targetPath := filepath.Join(target, relPath)
			validationLevel := s.determineValidationLevel(relPath)

			comp := &FileComparison{
				Path:   relPath,
				Source: file,
				Level:  types.ValidationLevel(validationLevel),
			}

			if tf := findFile(targetStats, targetPath); tf != nil {
				comp.Target = tf
				comp.Status = s.compareFiles(file, tf, comp.Level)
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
				logger.Debug("Error getting relative path for target file",
					"path", file.Path,
					"error", err)
				continue
			}

			if matchesPattern(relPath, s.opts.ExcludePatterns) {
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
	logger.Debug("Using default validation level",
		"path", path,
		"level", s.opts.Level)
	return s.opts.Level
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
	const modTimeToleranceSeconds = 2

	// Metadata check for all levels
	if src.Size != tgt.Size {
		logger.Debug("Size mismatch",
			"path", src.Path,
			"sourceSize", src.Size,
			"targetSize", tgt.Size)
		return StatusDiffer
	}

	if diff := abs(src.ModTime - tgt.ModTime); diff > modTimeToleranceSeconds {
		logger.Debug("Modification time mismatch",
			"path", src.Path,
			"sourceTime", src.ModTime,
			"targetTime", tgt.ModTime)
		return StatusDiffer
	}

	// Quick validation stops at metadata
	if level == types.Quick {
		return StatusMatch
	}

	// Standard validation: 32KB hash
	if level == types.Standard {
		bufferSize := s.opts.ValidationConfig.BufferSize
		if bufferSize == 0 {
			bufferSize = 32768
		}

		srcHash, err := hashPartialFile(src.Path, bufferSize)
		if err != nil {
			logger.Debug("Hash error", "path", src.Path, "error", err)
			return StatusError
		}
		tgtHash, err := hashPartialFile(tgt.Path, bufferSize)
		if err != nil {
			logger.Debug("Hash error", "path", tgt.Path, "error", err)
			return StatusError
		}

		if srcHash != tgtHash {
			logger.Debug("Content mismatch",
				"path", src.Path,
				"level", level)
			return StatusDiffer
		}
		return StatusMatch
	}

	// Deep validation: full content hash
	srcHash, err := hashFile(src.Path)
	if err != nil {
		logger.Debug("Hash error", "path", src.Path, "error", err)
		return StatusError
	}
	tgtHash, err := hashFile(tgt.Path)
	if err != nil {
		logger.Debug("Hash error", "path", tgt.Path, "error", err)
		return StatusError
	}

	if srcHash != tgtHash {
		logger.Debug("Content mismatch",
			"path", src.Path,
			"level", level)
		return StatusDiffer
	}
	return StatusMatch
}

// hashPartialFile reads and hashes only the first bufferSize bytes
func hashPartialFile(path string, bufferSize int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	buf := make([]byte, bufferSize)

	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}

	h.Write(buf[:n])
	return string(h.Sum(nil)), nil
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
