// internal/validation/deep.go
package validation

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/scan"
	"github.com/jack-sneddon/backup-butler/internal/types"
)

// DeepValidator implements full content comparison
type DeepValidator struct {
	opts *ValidatorOptions
}

func NewDeepValidator(opts *ValidatorOptions) *DeepValidator {
	if opts == nil {
		opts = &ValidatorOptions{
			BufferSize: 1024 * 1024, // 1MB default for deep validation
			Algorithm:  "sha256",
		}
	}
	return &DeepValidator{opts: opts}
}

func (v *DeepValidator) Level() types.ValidationLevel {
	return types.Deep
}

func (v *DeepValidator) Compare(source, target *scan.FileInfo) ComparisonResult {
	start := time.Now()

	// First do quick comparison
	quickResult := NewQuickValidator().Compare(source, target)
	if !quickResult.Equal {
		return quickResult
	}

	// Open both files
	srcFile, err := os.Open(source.Path)
	if err != nil {
		return ComparisonResult{
			Equal:     false,
			Reason:    fmt.Sprintf("Error opening source file: %v", err),
			TimeTaken: time.Since(start),
		}
	}
	defer srcFile.Close()

	tgtFile, err := os.Open(target.Path)
	if err != nil {
		return ComparisonResult{
			Equal:     false,
			Reason:    fmt.Sprintf("Error opening target file: %v", err),
			TimeTaken: time.Since(start),
		}
	}
	defer tgtFile.Close()

	// Compare content in chunks
	srcBuf := make([]byte, v.opts.BufferSize)
	tgtBuf := make([]byte, v.opts.BufferSize)
	var bytesRead int64

	for {
		srcN, srcErr := srcFile.Read(srcBuf)
		tgtN, tgtErr := tgtFile.Read(tgtBuf)
		bytesRead += int64(srcN)

		// Check for read errors
		if srcErr != nil && srcErr != io.EOF {
			return ComparisonResult{
				Equal:     false,
				Reason:    fmt.Sprintf("Error reading source file: %v", srcErr),
				TimeTaken: time.Since(start),
				BytesRead: bytesRead,
			}
		}
		if tgtErr != nil && tgtErr != io.EOF {
			return ComparisonResult{
				Equal:     false,
				Reason:    fmt.Sprintf("Error reading target file: %v", tgtErr),
				TimeTaken: time.Since(start),
				BytesRead: bytesRead,
			}
		}

		// Check for EOF
		if srcErr == io.EOF && tgtErr == io.EOF {
			break
		}
		if srcErr == io.EOF || tgtErr == io.EOF {
			return ComparisonResult{
				Equal:     false,
				Reason:    "Files have different lengths",
				TimeTaken: time.Since(start),
				BytesRead: bytesRead,
			}
		}

		// Compare chunks
		if srcN != tgtN {
			return ComparisonResult{
				Equal:     false,
				Reason:    "Files have different lengths",
				TimeTaken: time.Since(start),
				BytesRead: bytesRead,
			}
		}

		// Compare content
		if !bytesEqual(srcBuf[:srcN], tgtBuf[:tgtN]) {
			return ComparisonResult{
				Equal:     false,
				Reason:    "Content differs",
				TimeTaken: time.Since(start),
				BytesRead: bytesRead,
			}
		}
	}

	return ComparisonResult{
		Equal:     true,
		Reason:    "Full content match",
		TimeTaken: time.Since(start),
		BytesRead: bytesRead,
	}
}

// bytesEqual performs a constant-time comparison of two byte slices
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
