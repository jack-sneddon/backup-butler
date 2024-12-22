// internal/adapters/storage/filesystem/adapter.go
package filesystem

import (
	"github.com/jack-sneddon/backup-butler/internal/core/storage"
	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type FilesystemAdapter struct {
	manager *storage.Manager
}

func NewFilesystemAdapter(checksumAlg string, bufferSize int) *FilesystemAdapter {
	// Create the core storage manager
	manager := storage.NewManager(bufferSize)

	return &FilesystemAdapter{
		manager: manager,
	}
}

// Implement backup.StoragePort interface using our core Manager
func (a *FilesystemAdapter) CalculateChecksum(filePath string) (string, error) {
	return a.manager.CalculateChecksum(filePath)
}

func (a *FilesystemAdapter) Copy(src, dst string, bufferSize int) (int64, error) {
	return a.manager.Copy(src, dst, bufferSize)
}

func (a *FilesystemAdapter) Exists(path string) (bool, error) {
	return a.manager.Exists(path)
}

func (a *FilesystemAdapter) GetMetadata(path string) (backup.FileMetadata, error) {
	return a.manager.GetMetadata(path)
}

func (a *FilesystemAdapter) CreateDirectory(path string) error {
	return a.manager.CreateDirectory(path)
}

func (a *FilesystemAdapter) IsDirectory(path string) (bool, error) {
	return a.manager.IsDirectory(path)
}
