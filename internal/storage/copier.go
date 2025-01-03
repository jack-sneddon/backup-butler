// internal/storage/copier.go
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Copier handles file copy operations with verification
type Copier struct {
	manager    *Manager
	bufferSize int
}

// NewCopier creates a new copier instance
func NewCopier(manager *Manager, bufferSize int) *Copier {
	return &Copier{
		manager:    manager,
		bufferSize: bufferSize,
	}
}

// Copy performs the file copy operation with verification
func (c *Copier) Copy(ctx context.Context, src, dst string) (CopyResult, error) {
	startTime := time.Now()

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return CopyResult{}, fmt.Errorf("failed to create directory: %w", err)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return CopyResult{}, fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return CopyResult{}, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Prepare buffer for copying
	buffer := make([]byte, c.bufferSize)

	// Copy with progress tracking
	var written int64
	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return CopyResult{
				Source:      src,
				Destination: dst,
				BytesCopied: written,
				Duration:    time.Since(startTime),
				Error:       ctx.Err(),
			}, ctx.Err()
		default:
		}

		nr, err := srcFile.Read(buffer)
		if nr > 0 {
			nw, err := dstFile.Write(buffer[0:nr])
			if err != nil {
				return CopyResult{}, fmt.Errorf("write error: %w", err)
			}
			if nr != nw {
				return CopyResult{}, fmt.Errorf("short write")
			}
			written += int64(nw)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return CopyResult{}, fmt.Errorf("read error: %w", err)
		}
	}

	// Sync to ensure data is written to disk
	if err := dstFile.Sync(); err != nil {
		return CopyResult{}, fmt.Errorf("failed to sync file: %w", err)
	}

	// Copy file mode and timestamps
	srcInfo, err := os.Stat(src)
	if err != nil {
		return CopyResult{}, fmt.Errorf("failed to get source info: %w", err)
	}

	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return CopyResult{}, fmt.Errorf("failed to set permissions: %w", err)
	}

	// Copy file mode and timestamps (add this after chmod)
	if err := os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		return CopyResult{}, fmt.Errorf("failed to set timestamps: %w", err)
	}

	return CopyResult{
		Source:      src,
		Destination: dst,
		BytesCopied: written,
		Duration:    time.Since(startTime),
	}, nil
}

// VerifyCopy verifies the integrity of a copied file
func (c *Copier) VerifyCopy(src, dst string) error {
	srcChecksum, err := calculateFullChecksum(src)
	if err != nil {
		return fmt.Errorf("failed to calculate source checksum: %w", err)
	}

	dstChecksum, err := calculateFullChecksum(dst)
	if err != nil {
		return fmt.Errorf("failed to calculate destination checksum: %w", err)
	}

	if srcChecksum != dstChecksum {
		return fmt.Errorf("checksum mismatch: copy may be corrupted")
	}

	return nil
}
