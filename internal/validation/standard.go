// internal/validation/standard.go
package validation

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/scan"
	"github.com/jack-sneddon/backup-butler/internal/types"
)

// StandardValidator implements partial content comparison
type StandardValidator struct {
	opts *ValidatorOptions
}

func NewStandardValidator(opts *ValidatorOptions) *StandardValidator {
	if opts == nil {
		opts = &ValidatorOptions{
			BufferSize: 32768, // 32KB default
			Algorithm:  "sha256",
		}
	}
	return &StandardValidator{opts: opts}
}

func (v *StandardValidator) Level() types.ValidationLevel {
	return types.Standard
}

func (v *StandardValidator) Compare(source, target *scan.FileInfo) ComparisonResult {
	start := time.Now()

	// First do quick comparison
	quickResult := NewQuickValidator().Compare(source, target)
	if !quickResult.Equal {
		return quickResult
	}

	// Initialize hash function based on configuration
	hashFunc := v.getHashFunc()
	if hashFunc == nil {
		return ComparisonResult{
			Equal:     false,
			Reason:    fmt.Sprintf("Unsupported hash algorithm: %s", v.opts.Algorithm),
			TimeTaken: time.Since(start),
		}
	}

	// Read and hash the first buffer of each file
	sourceHash, bytesRead, err := v.hashFileStart(source.Path, hashFunc)
	if err != nil {
		return ComparisonResult{
			Equal:     false,
			Reason:    fmt.Sprintf("Error reading source file: %v", err),
			TimeTaken: time.Since(start),
			BytesRead: bytesRead,
		}
	}

	// Reset hash function for target file
	hashFunc = v.getHashFunc()
	targetHash, targetBytesRead, err := v.hashFileStart(target.Path, hashFunc)
	if err != nil {
		return ComparisonResult{
			Equal:     false,
			Reason:    fmt.Sprintf("Error reading target file: %v", err),
			TimeTaken: time.Since(start),
			BytesRead: bytesRead + targetBytesRead,
		}
	}

	return ComparisonResult{
		Equal:     sourceHash == targetHash,
		Reason:    v.getComparisonReason(sourceHash == targetHash),
		TimeTaken: time.Since(start),
		BytesRead: bytesRead + targetBytesRead,
	}
}

func (v *StandardValidator) getHashFunc() hash.Hash {
	switch v.opts.Algorithm {
	case "md5":
		return md5.New()
	case "sha1":
		return sha1.New()
	case "sha256":
		return sha256.New()
	default:
		return nil
	}
}

func (v *StandardValidator) hashFileStart(path string, h hash.Hash) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	buf := make([]byte, v.opts.BufferSize)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", int64(n), err
	}

	h.Write(buf[:n])
	return fmt.Sprintf("%x", h.Sum(nil)), int64(n), nil
}

func (v *StandardValidator) getComparisonReason(equal bool) string {
	if equal {
		return "Content match (first 32KB)"
	}
	return "Content differs"
}
