// internal/scan/compare.go
package scan

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
)

// internal/scan/compare.go
func (s *Scanner) Compare(source, target string) ([]*FileComparison, error) {
	// Reset stats before scanning
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

	// Compare source files to target
	for _, dir := range sourceStats {
		for _, file := range dir.Files {
			relPath, _ := filepath.Rel(source, file.Path)
			if relPath, _ := filepath.Rel(source, file.Path); relPath == "test_config.yaml" {
				continue
			}

			targetPath := filepath.Join(target, relPath)

			comp := &FileComparison{
				Path:   relPath,
				Source: file,
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
			relPath, _ := filepath.Rel(target, file.Path)
			if relPath == "test_config.yaml" {
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
		s.log.Debugw("Size mismatch",
			"path", src.Path,
			"srcSize", src.Size,
			"tgtSize", tgt.Size)
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
