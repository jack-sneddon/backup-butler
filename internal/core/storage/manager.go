// internal/core/storage/manager.go
package storage

import (
	"io"
	"os"
	"path/filepath"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type Manager struct {
	checksumCalculator *ChecksumCalculator
	bufferSize         int
}

// NewManager creates a new storage manager
func NewManager(bufferSize int) *Manager {
	if bufferSize <= 0 {
		bufferSize = 32 * 1024 // 32KB default
	}
	return &Manager{
		checksumCalculator: NewChecksumCalculator(),
		bufferSize:         bufferSize,
	}
}

// CalculateChecksum calculates the checksum of a file
func (m *Manager) CalculateChecksum(filePath string) (string, error) {
	checksum, err := m.checksumCalculator.CalculateChecksum(filePath)
	if err != nil {
		return "", newStorageError(ErrChecksumMismatch, "CalculateChecksum", filePath, err)
	}
	return checksum, nil
}

// Copy copies a file with the specified buffer size
func (m *Manager) Copy(src, dst string, bufferSize int) (int64, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, newStorageError(ErrAccessDenied, "OpenFile", src, err)
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return 0, newStorageError(ErrDirectoryCreationFailed, "CreateDirectory", dst, err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return 0, newStorageError(ErrAccessDenied, "CreateFile", dst, err)
	}
	defer dstFile.Close()

	if bufferSize <= 0 {
		bufferSize = m.bufferSize
	}
	buffer := make([]byte, bufferSize)

	copied, err := io.CopyBuffer(dstFile, srcFile, buffer)
	if err != nil {
		return copied, newCopyError(src, dst, err)
	}

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return copied, newStorageError(ErrMetadataReadFailed, "GetFileInfo", src, err)
	}

	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return copied, newStorageError(ErrAccessDenied, "SetPermissions", dst, err)
	}

	return copied, nil
}

// Exists checks if a file or directory exists
func (m *Manager) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, newStorageError(ErrAccessDenied, "CheckExists", path, err)
}

// GetMetadata gets file metadata
func (m *Manager) GetMetadata(path string) (backup.FileMetadata, error) {
	info, err := os.Stat(path)
	if err != nil {
		return backup.FileMetadata{}, newStorageError(ErrMetadataReadFailed, "GetMetadata", path, err)
	}

	metadata := backup.FileMetadata{
		Path:    path,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}

	if !info.IsDir() {
		checksum, err := m.CalculateChecksum(path)
		if err != nil {
			return metadata, err // Error already wrapped by CalculateChecksum
		}
		metadata.Checksum = checksum
	}

	return metadata, nil
}

// CreateDirectory creates a directory and any necessary parents
func (m *Manager) CreateDirectory(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return newStorageError(ErrDirectoryCreationFailed, "CreateDirectory", path, err)
	}
	return nil
}

// IsDirectory checks if the path is a directory
func (m *Manager) IsDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, newStorageError(ErrFileNotFound, "IsDirectory", path, err)
		}
		return false, newStorageError(ErrAccessDenied, "IsDirectory", path, err)
	}
	return info.IsDir(), nil
}
