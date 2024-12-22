// internal/core/storage/errors.go
package storage

import "fmt"

type StorageErrorCode int

const (
	ErrUnknown StorageErrorCode = iota
	ErrFileNotFound
	ErrAccessDenied
	ErrInvalidPath
	ErrChecksumMismatch
	ErrSizeMismatch
	ErrCopyFailed
	ErrDirectoryCreationFailed
	ErrMetadataReadFailed
	ErrInvalidOperation
)

type StorageError struct {
	Code  StorageErrorCode
	Op    string
	Path  string
	Path2 string
	Err   error
}

func (e *StorageError) Error() string {
	switch {
	case e.Path2 != "":
		return fmt.Sprintf("%s failed: %s -> %s: %v", e.Op, e.Path, e.Path2, e.Err)
	case e.Path != "":
		return fmt.Sprintf("%s failed: %s: %v", e.Op, e.Path, e.Err)
	default:
		return fmt.Sprintf("%s failed: %v", e.Op, e.Err)
	}
}

func (e *StorageError) Is(target error) bool {
	t, ok := target.(*StorageError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func (e *StorageError) Unwrap() error {
	return e.Err
}

func newStorageError(code StorageErrorCode, op, path string, err error) *StorageError {
	return &StorageError{
		Code: code,
		Op:   op,
		Path: path,
		Err:  err,
	}
}

func newCopyError(src, dst string, err error) *StorageError {
	return &StorageError{
		Code:  ErrCopyFailed,
		Op:    "Copy",
		Path:  src,
		Path2: dst,
		Err:   err,
	}
}
