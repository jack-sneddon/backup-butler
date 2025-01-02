// internal/storage/manager.go
package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/types"
)

/*
const (
	QuickHashSize = 64 * 1024 // 64KB for quick hash check
)

type ComparisonStrategy int

const (
	CompareMetadataOnly ComparisonStrategy = iota
	CompareQuickHash
	CompareFullChecksum
)
*/

// Manager handles file operations
type Manager struct {
	baseDir            string
	bufferSize         int
	checksumCalculator *ChecksumCalculator
}

// NewManager creates a new storage manager
func NewManager(targetDir string, bufferSize int) *Manager {
	if bufferSize <= 0 {
		bufferSize = 32 * 1024 // 32KB default
	}
	return &Manager{
		baseDir:            targetDir,
		bufferSize:         bufferSize,
		checksumCalculator: NewChecksumCalculator(),
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

	checksum, err := m.checksumCalculator.CalculateChecksum(path)
	if err != nil {
		return metadata, fmt.Errorf("failed to calculate checksum: %w", err)
	}
	metadata.Checksum = checksum

	return metadata, nil
}

func (m *Manager) calculateQuickHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	buffer := make([]byte, QuickHashSize)

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	hash.Write(buffer[:n])
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Compare determines if a file needs to be copied
func (m *Manager) Compare(src, dst string, versionMgr types.VersionManager) (CompareResult, error) {
	srcInfo, err := m.GetMetadata(src)
	if err != nil {
		return CompareResult{
			NeedsCopy: true,
			Reason:    "failed to read source metadata",
			Strategy:  CompareMetadataOnly,
		}, fmt.Errorf("source read error: %w", err)
	}

	// Check integrity against last version
	lastVersion, _ := versionMgr.GetFileLastVersion(src)
	if check := checkFileIntegrity(src, srcInfo, lastVersion); check != nil {
		// Store integrity check result but continue with normal comparison
		m.reportIntegrityIssue(check)
	}

	// Get last known good version info
	/*
	   lastVersion, err := versionMgr.GetFileLastVersion(src)
	   if err == nil {
	       // File exists in version history, check for potential corruption
	       if srcInfo.ModTime.Equal(lastVersion.ModTime) {
	           // File hasn't been modified, but...
	           if srcInfo.Size != lastVersion.Size {
	               return CompareResult{
	                   NeedsCopy: true,
	                   Reason:    "possible corruption: size changed without modification",
	                   Strategy:  CompareMetadataOnly,
	                   Source:    srcInfo,
	                   Warnings:  []string{"File size changed without modification time change"},
	               }, nil
	           }
	       }
	   }
	*/

	// Check destination
	dstInfo, err := m.GetMetadata(dst)
	if err != nil {
		if os.IsNotExist(err) {
			return CompareResult{
				NeedsCopy: true,
				Reason:    "destination does not exist",
				Strategy:  CompareMetadataOnly,
				Source:    srcInfo,
			}, nil
		}
		return CompareResult{
			NeedsCopy: true,
			Reason:    "destination read error",
			Strategy:  CompareMetadataOnly,
			Source:    srcInfo,
		}, nil
	}

	// Size comparison with corruption check
	if srcInfo.Size != dstInfo.Size {
		if lastVersion != nil && dstInfo.Size == lastVersion.Size {
			return CompareResult{
				NeedsCopy: true,
				Reason:    "possible source corruption: size differs from last known good copy",
				Strategy:  CompareMetadataOnly,
				Source:    srcInfo,
				Target:    dstInfo,
				Warnings:  []string{"Source file size differs from last known good backup"},
			}, nil
		}
		return CompareResult{
			NeedsCopy: true,
			Reason:    "size mismatch",
			Strategy:  CompareMetadataOnly,
			Source:    srcInfo,
			Target:    dstInfo,
		}, nil
	}

	// Quick hash with corruption detection
	srcQuickHash, err := m.calculateQuickHash(src)
	if err == nil {
		if lastVersion != nil && lastVersion.QuickHash != "" {
			if srcInfo.ModTime.Equal(lastVersion.ModTime) && srcQuickHash != lastVersion.QuickHash {
				return CompareResult{
					NeedsCopy: true,
					Reason:    "possible corruption: content changed without modification",
					Strategy:  CompareQuickHash,
					Source:    srcInfo,
					Target:    dstInfo,
					Warnings:  []string{"File content changed without modification time change"},
				}, nil
			}
		}

		dstQuickHash, err := m.calculateQuickHash(dst)
		if err == nil && srcQuickHash == dstQuickHash {
			return CompareResult{
				NeedsCopy: false,
				Reason:    "quick hash match",
				Strategy:  CompareQuickHash,
				Source:    srcInfo,
				Target:    dstInfo,
				QuickHash: srcQuickHash,
			}, nil
		}
	}

	// Full checksum with corruption detection
	if lastVersion != nil && srcInfo.ModTime.Equal(lastVersion.ModTime) && srcInfo.Checksum != lastVersion.Checksum {
		return CompareResult{
			NeedsCopy: true,
			Reason:    "possible corruption: checksum changed without modification",
			Strategy:  CompareFullChecksum,
			Source:    srcInfo,
			Target:    dstInfo,
			Warnings:  []string{"File checksum changed without modification time change"},
		}, nil
	}

	// Normal checksum comparison
	if srcInfo.Checksum == dstInfo.Checksum {
		return CompareResult{
			NeedsCopy: false,
			Reason:    "checksum match",
			Strategy:  CompareFullChecksum,
			Source:    srcInfo,
			Target:    dstInfo,
		}, nil
	}

	return CompareResult{
		NeedsCopy: true,
		Reason:    "checksum mismatch",
		Strategy:  CompareFullChecksum,
		Source:    srcInfo,
		Target:    dstInfo,
	}, nil
}

// internal/storage/manager.go
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

	return CopyResult{
		Source:      src,
		Destination: dst,
		BytesCopied: copied,
		Duration:    time.Since(startTime),
	}, nil
}

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
