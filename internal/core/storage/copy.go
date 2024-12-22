// internal/core/storage/copy.go
package storage

import (
	"context"
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
	"time"
)

type CopyOperation struct {
	Source      string
	Destination string
	Size        int64
	ModTime     time.Time
	Checksum    string
	VerifyMode  bool
}

type CopyResult struct {
	Operation CopyOperation
	Copied    int64
	Duration  time.Duration
	Error     error
}

func (m *Manager) CopyFile(ctx context.Context, op CopyOperation) CopyResult {
	startTime := time.Now()
	result := CopyResult{Operation: op}

	srcFile, err := os.Open(op.Source)
	if err != nil {
		result.Error = newStorageError(ErrAccessDenied, "OpenFile", op.Source, err)
		return result
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(op.Destination), 0755); err != nil {
		result.Error = newStorageError(ErrDirectoryCreationFailed, "CreateDirectory", op.Destination, err)
		return result
	}

	dstFile, err := os.Create(op.Destination)
	if err != nil {
		result.Error = newStorageError(ErrAccessDenied, "CreateFile", op.Destination, err)
		return result
	}
	defer dstFile.Close()

	buffer := make([]byte, m.bufferSize)
	var writer io.Writer = dstFile
	hasher := sha256.New()
	if op.VerifyMode {
		writer = io.MultiWriter(dstFile, hasher)
	}

	copied, err := m.copyWithContext(ctx, writer, srcFile, buffer)
	if err != nil {
		result.Error = newCopyError(op.Source, op.Destination, err)
		return result
	}
	result.Copied = copied

	if err := m.preserveFileAttributes(op.Source, op.Destination); err != nil {
		result.Error = newStorageError(ErrAccessDenied, "SetAttributes", op.Destination, err)
		return result
	}

	result.Duration = time.Since(startTime)
	return result
}

func (m *Manager) copyWithContext(ctx context.Context, dst io.Writer, src io.Reader, buf []byte) (int64, error) {
	var written int64

	for {
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		default:
			nr, readErr := src.Read(buf)
			if nr > 0 {
				nw, writeErr := dst.Write(buf[0:nr])
				if nw < 0 || nr < nw {
					nw = 0
					if writeErr == nil {
						writeErr = newStorageError(ErrCopyFailed, "Write", "", nil)
					}
				}
				written += int64(nw)
				if writeErr != nil {
					return written, writeErr
				}
				if nr != nw {
					return written, io.ErrShortWrite
				}
			}
			if readErr != nil {
				if readErr == io.EOF {
					return written, nil
				}
				return written, readErr
			}
		}
	}
}

func (m *Manager) preserveFileAttributes(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return newStorageError(ErrMetadataReadFailed, "GetFileInfo", src, err)
	}

	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return newStorageError(ErrAccessDenied, "SetPermissions", dst, err)
	}

	return nil
}
