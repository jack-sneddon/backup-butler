// internal/scan/compare.go
package scan

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
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
				Level:  validationLevel,
			}

			if tf := findFile(targetStats, targetPath); tf != nil {
				comp.Target = tf
				comp.Status = s.compareFiles(file, tf)
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

func (s *Scanner) determineValidationLevel(path string) string {
	if s.opts.ValidationConfig != nil && s.opts.ValidationConfig.CriticalPaths != nil {
		for _, cp := range s.opts.ValidationConfig.CriticalPaths {
			matched, err := filepath.Match(cp.Path, path) // Changed from Pattern to Path
			if err == nil && matched {
				s.log.Debugw("Using critical path validation level",
					"path", path,
					"pattern", cp.Path, // Changed from Pattern to Path
					"level", cp.Level)
				return cp.Level
			}
		}
	}
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

func (s *Scanner) compareFiles(src, tgt *FileInfo) FileStatus {
	if src.Size != tgt.Size {
		// If files differ and we have validation config with onMismatch
		if s.opts.ValidationConfig != nil && s.opts.ValidationConfig.OnMismatch != "" {
			s.log.Debugw("Escalating validation due to mismatch",
				"path", src.Path,
				"escalatedLevel", s.opts.ValidationConfig.OnMismatch)
			src.Level = s.opts.ValidationConfig.OnMismatch
		}
		return StatusDiffer
	}
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

	if srcHash == tgtHash {
		return StatusMatch
	}
	return StatusDiffer
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

func (s *Scanner) findTargetFile(sourcePath, targetRoot string) *FileInfo {
	relPath, _ := filepath.Rel(s.rootPath, sourcePath)
	targetPath := filepath.Join(targetRoot, relPath)

	for _, dir := range s.stats {
		for _, file := range dir.Files {
			if file.Path == targetPath {
				return file
			}
		}
	}
	return nil
}
