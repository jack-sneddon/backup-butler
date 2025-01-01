// internal/storage/manager.go
package storage

import (
	"crypto/sha256"
	"encoding/hex"
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
	bufferSize         int
	checksumCalculator *ChecksumCalculator
}

// NewManager creates a new storage manager
func NewManager(bufferSize int) *Manager {
	if bufferSize <= 0 {
		bufferSize = 32 * 1024 // 32KB default
	}
	return &Manager{
		bufferSize:         bufferSize,
		checksumCalculator: NewChecksumCalculator(),
	}
}

// GetMetadata retrieves file metadata including checksum
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

// Compare determines if a file needs to be copied
func (m *Manager) Compare(src, dst string, versionMgr types.VersionManager) (CompareResult, error) {
	// Get source metadata (always needed)
	srcInfo, err := m.GetMetadata(src)
	if err != nil {
		return CompareResult{
			NeedsCopy: true,
			Reason:    "failed to read source metadata",
			Strategy:  CompareMetadataOnly,
		}, fmt.Errorf("source read error: %w", err)
	}

	// Check if destination exists
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
			Reason:    "failed to read destination metadata",
			Strategy:  CompareMetadataOnly,
			Source:    srcInfo,
		}, fmt.Errorf("destination read error: %w", err)
	}

	// Get last known version info if available
	lastVersion, err := versionMgr.GetFileLastVersion(src)
	if err == nil {
		// We have version history
		if srcInfo.Size != lastVersion.Size {
			return CompareResult{
				NeedsCopy:   true,
				Reason:      "size different from last backup",
				Strategy:    CompareMetadataOnly,
				Source:      srcInfo,
				Target:      dstInfo,
				LastVersion: lastVersion.ID,
			}, nil
		}

		if srcInfo.ModTime.Before(lastVersion.ModTime) {
			return CompareResult{
				NeedsCopy:   false,
				Reason:      "file older than last backup",
				Strategy:    CompareMetadataOnly,
				Source:      srcInfo,
				Target:      dstInfo,
				LastVersion: lastVersion.ID,
			}, nil
		}

		// File is newer, do quick hash check
		srcQuickHash, err := m.calculateQuickHash(src)
		if err != nil {
			// Fall back to metadata only if quick hash fails
			return CompareResult{
				NeedsCopy: true,
				Reason:    "quick hash failed",
				Strategy:  CompareMetadataOnly,
				Source:    srcInfo,
				Target:    dstInfo,
			}, nil
		}

		if srcQuickHash == lastVersion.QuickHash {
			return CompareResult{
				NeedsCopy:   false,
				Reason:      "quick hash matches last backup",
				Strategy:    CompareQuickHash,
				Source:      srcInfo,
				Target:      dstInfo,
				QuickHash:   srcQuickHash,
				LastVersion: lastVersion.ID,
			}, nil
		}
	}

	// No version history or quick hash mismatch - compare current files
	if srcInfo.Size != dstInfo.Size {
		return CompareResult{
			NeedsCopy: true,
			Reason:    "size mismatch",
			Strategy:  CompareMetadataOnly,
			Source:    srcInfo,
			Target:    dstInfo,
		}, nil
	}

	// If we get here, we need to do a full checksum
	srcChecksum, err := m.checksumCalculator.CalculateChecksum(src)
	if err != nil {
		return CompareResult{
			NeedsCopy: true,
			Reason:    "failed to calculate source checksum",
			Strategy:  CompareFullChecksum,
			Source:    srcInfo,
			Target:    dstInfo,
		}, nil
	}
	// dstChecksum, err := m.CalculateChecksum(dst)
	dstChecksum, err := m.checksumCalculator.CalculateChecksum(dst)
	if err != nil {
		return CompareResult{
			NeedsCopy: true,
			Reason:    "failed to calculate destination checksum",
			Strategy:  CompareFullChecksum,
			Source:    srcInfo,
			Target:    dstInfo,
		}, nil
	}

	if srcChecksum != dstChecksum {
		return CompareResult{
			NeedsCopy: true,
			Reason:    "checksum mismatch",
			Strategy:  CompareFullChecksum,
			Source:    srcInfo,
			Target:    dstInfo,
		}, nil
	}

	return CompareResult{
		NeedsCopy: false,
		Reason:    "files match",
		Strategy:  CompareFullChecksum,
		Source:    srcInfo,
		Target:    dstInfo,
	}, nil
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
