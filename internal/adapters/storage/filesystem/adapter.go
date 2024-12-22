package filesystem

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

type FilesystemAdapter struct {
	checksumAlgorithm string
}

func NewFilesystemAdapter(checksumAlg string) *FilesystemAdapter {
	if checksumAlg == "" {
		checksumAlg = "sha256"
	}
	return &FilesystemAdapter{
		checksumAlgorithm: checksumAlg,
	}
}

func (a *FilesystemAdapter) CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (a *FilesystemAdapter) Copy(src, dst string, bufferSize int) (int64, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return 0, fmt.Errorf("failed to create destination directory: %w", err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return 0, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	buf := make([]byte, bufferSize)
	copied, err := io.CopyBuffer(dstFile, srcFile, buf)
	if err != nil {
		return copied, fmt.Errorf("failed during copy: %w", err)
	}

	// Preserve file mode
	srcInfo, err := srcFile.Stat()
	if err == nil {
		if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
			return copied, fmt.Errorf("failed to preserve file mode: %w", err)
		}
	}

	return copied, nil
}

func (a *FilesystemAdapter) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (a *FilesystemAdapter) GetMetadata(path string) (backup.FileMetadata, error) {
	info, err := os.Stat(path)
	if err != nil {
		return backup.FileMetadata{}, fmt.Errorf("failed to get file info: %w", err)
	}

	checksum := ""
	if !info.IsDir() {
		checksum, err = a.CalculateChecksum(path)
		if err != nil {
			return backup.FileMetadata{}, fmt.Errorf("failed to calculate checksum: %w", err)
		}
	}

	return backup.FileMetadata{
		Path:     path,
		Size:     info.Size(),
		ModTime:  info.ModTime(),
		Checksum: checksum,
	}, nil
}

func (a *FilesystemAdapter) CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

func (a *FilesystemAdapter) IsDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
