package mock

import (
	"time"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type MockStorageAdapter struct {
	CalculateChecksumFunc func(filePath string) (string, error)
	CopyFunc              func(src, dst string, bufferSize int) (int64, error)
	ExistsFunc            func(path string) (bool, error)
	GetMetadataFunc       func(path string) (backup.FileMetadata, error)
	CreateDirectoryFunc   func(path string) error
	IsDirectoryFunc       func(path string) (bool, error)
}

func NewMockStorageAdapter() *MockStorageAdapter {
	return &MockStorageAdapter{
		CalculateChecksumFunc: func(filePath string) (string, error) {
			return "mock-checksum", nil
		},
		CopyFunc: func(src, dst string, bufferSize int) (int64, error) {
			return 1024, nil
		},
		ExistsFunc: func(path string) (bool, error) {
			return true, nil
		},
		GetMetadataFunc: func(path string) (backup.FileMetadata, error) {
			return backup.FileMetadata{
				Path:     path,
				Size:     1024,
				ModTime:  time.Now(),
				Checksum: "mock-checksum",
			}, nil
		},
		CreateDirectoryFunc: func(path string) error {
			return nil
		},
		IsDirectoryFunc: func(path string) (bool, error) {
			return false, nil
		},
	}
}

// Implement all interface methods using the function fields
func (m *MockStorageAdapter) CalculateChecksum(filePath string) (string, error) {
	return m.CalculateChecksumFunc(filePath)
}

func (m *MockStorageAdapter) Copy(src, dst string, bufferSize int) (int64, error) {
	return m.CopyFunc(src, dst, bufferSize)
}

func (m *MockStorageAdapter) Exists(path string) (bool, error) {
	return m.ExistsFunc(path)
}

func (m *MockStorageAdapter) GetMetadata(path string) (backup.FileMetadata, error) {
	return m.GetMetadataFunc(path)
}

func (m *MockStorageAdapter) CreateDirectory(path string) error {
	return m.CreateDirectoryFunc(path)
}

func (m *MockStorageAdapter) IsDirectory(path string) (bool, error) {
	return m.IsDirectoryFunc(path)
}
