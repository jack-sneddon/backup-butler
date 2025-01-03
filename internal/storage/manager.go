// internal/storage/manager.go
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jack-sneddon/backup-butler/internal/config"
	"github.com/jack-sneddon/backup-butler/internal/types"
)

// Manager handles file operations
type Manager struct {
	baseDir    string
	bufferSize int
	config     *config.Config // Add this field
}

// NewManager creates a new storage manager
func NewManager(targetDir string, bufferSize int, cfg *config.Config) *Manager {
	if bufferSize <= 0 {
		bufferSize = 32 * 1024 // 32KB default
	}
	return &Manager{
		baseDir:    targetDir,
		bufferSize: bufferSize,
		config:     cfg,
	}
}

// GetMetadata retrieves file metadata including checksum
func (m *Manager) GetMetadata(path string) (types.FileMetadata, error) {
	info, err := os.Stat(path)
	if err != nil {
		return types.FileMetadata{}, fmt.Errorf("failed to get file info: %w", err)
	}

	metadata := types.FileMetadata{
		Path:    path,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}

	checksum, err := calculateFullChecksum(path)
	if err != nil {
		return metadata, fmt.Errorf("failed to calculate checksum: %w", err)
	}
	metadata.Checksum = checksum

	return metadata, nil
}

// Compare determines if a file needs to be copied
// internal/storage/manager.go
func (m *Manager) Compare(src, dst string, versionMgr types.VersionManager) (CompareResult, error) {
	if m.config.LogLevel >= config.LogDebug {
		relPath, _ := filepath.Rel(m.config.SourceDirectory, src)
		fmt.Printf("Comparing file: %s\n", relPath)
	}

	meta := &Metadata{}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return CompareResult{NeedsCopy: true, Reason: "source stat failed", Strategy: "metadata"}, err
	}
	meta.Size = srcInfo.Size()
	meta.ModTime = srcInfo.ModTime()

	strategies := []CompareStrategy{
		&MetadataCompare{},
		&QuickHashCompare{},
		&FullChecksumCompare{},
	}

	for _, strategy := range strategies {
		result, err := strategy.Compare(src, dst, meta)
		if err != nil {
			if m.config.LogLevel >= config.LogVerbose {
				fmt.Printf("Strategy %s failed for %s: %v\n",
					result.Strategy,
					filepath.Base(src),
					err)
			}
			continue
		}

		if m.config.LogLevel >= config.LogDebug {
			fmt.Printf("  Strategy: %s, Result: %v, Reason: %s\n",
				result.Strategy,
				!result.NeedsCopy,
				result.Reason)
		}

		if result.Reason != "try next strategy" {
			if m.config.LogLevel >= config.LogVerbose {
				fmt.Printf("%s: %s (%s)\n",
					filepath.Base(src),
					result.Reason,
					result.Strategy)
			}
			return result, nil
		}
	}

	return CompareResult{
		NeedsCopy: true,
		Reason:    "no strategy provided definitive answer",
		Strategy:  "fallback",
	}, nil
}

func (m *Manager) reportIntegrityIssue(check *IntegrityCheck) error {
	issuesPath := filepath.Join(m.baseDir, ".versions", "integrity_issues.json")

	var issues []*IntegrityCheck
	data, err := os.ReadFile(issuesPath)
	if err == nil {
		json.Unmarshal(data, &issues)
	}

	issues = append(issues, check)

	// Keep only recent issues (last 100)
	if len(issues) > 100 {
		issues = issues[len(issues)-100:]
	}

	data, err = json.MarshalIndent(issues, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal integrity issues: %w", err)
	}

	return os.WriteFile(issuesPath, data, 0644)
}

// CopyFile copies a file with buffer
/*
func (m *Manager) CopyFile(src, dst string) (CopyResult, error) {
	startTime := time.Now()

	srcFile, err := os.Open(src)
	if err != nil {
		return CopyResult{}, fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return CopyResult{}, fmt.Errorf("failed to create directory: %w", err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return CopyResult{}, fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()

	buffer := make([]byte, m.bufferSize)
	copied, err := io.CopyBuffer(dstFile, srcFile, buffer)
	if err != nil {
		return CopyResult{}, fmt.Errorf("failed to copy file: %w", err)
	}

	// Copy file mode and timestamps
	srcInfo, err := os.Stat(src)
	if err != nil {
		return CopyResult{}, fmt.Errorf("failed to get source info: %w", err)
	}

	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return CopyResult{}, fmt.Errorf("failed to set permissions: %w", err)
	}

	// Preserve timestamps
	if err := os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		return CopyResult{}, fmt.Errorf("failed to set timestamps: %w", err)
	}

	return CopyResult{
		Source:      src,
		Destination: dst,
		BytesCopied: copied,
		Duration:    time.Since(startTime),
	}, nil
}
*/

// Exists checks if a file or directory exists
func (m *Manager) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory checks if path is a directory
func (m *Manager) IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (m *Manager) GetIntegrityIssues() ([]*IntegrityCheck, error) {
	issuesPath := filepath.Join(m.baseDir, ".versions", "integrity_issues.json")

	data, err := os.ReadFile(issuesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read integrity issues: %w", err)
	}

	var issues []*IntegrityCheck
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, fmt.Errorf("failed to parse integrity issues: %w", err)
	}

	return issues, nil
}
