// internal/storage/compare.go
package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	QuickHashSize = 64 * 1024 // 64KB for quick hash
)

type CompareStrategy interface {
	Compare(src, dst string, meta *Metadata) (CompareResult, error)
	Priority() int // Lower number = higher priority
}

type Metadata struct {
	Size         int64
	ModTime      time.Time
	QuickHash    string
	FullChecksum string
}

// CompareResult now includes which strategy was used
type CompareResult struct {
	NeedsCopy bool
	Reason    string
	Strategy  string
}

// MetadataCompare is fastest but least thorough
type MetadataCompare struct{}

func (m *MetadataCompare) Compare(src, dst string, meta *Metadata) (CompareResult, error) {
	dstInfo, err := os.Stat(dst)
	if os.IsNotExist(err) {
		return CompareResult{NeedsCopy: true, Reason: "destination missing", Strategy: "metadata"}, nil
	}
	if err != nil {
		return CompareResult{NeedsCopy: true, Reason: "destination error", Strategy: "metadata"}, err
	}

	// Add debug output
	fmt.Printf("  Metadata check: src size=%d, dst size=%d\n", meta.Size, dstInfo.Size())
	fmt.Printf("  Metadata check: src time=%v, dst time=%v\n", meta.ModTime, dstInfo.ModTime())

	// First check - size
	if meta.Size != dstInfo.Size() {
		return CompareResult{NeedsCopy: true, Reason: "size mismatch", Strategy: "metadata"}, nil
	}

	// If size matches and timestamps match exactly, we can skip
	if meta.ModTime.Equal(dstInfo.ModTime()) {
		return CompareResult{NeedsCopy: false, Reason: "metadata exact match", Strategy: "metadata"}, nil
	}

	// If just timestamps are different, try quick hash
	return CompareResult{NeedsCopy: true, Reason: "try next strategy", Strategy: "metadata"}, nil
}

func (m *MetadataCompare) Priority() int {
	return 1
}

// QuickHashCompare calculates hash of first 64KB
type QuickHashCompare struct{}

func (q *QuickHashCompare) Compare(src, dst string, meta *Metadata) (CompareResult, error) {
	// Calculate source quick hash if not already done
	if meta.QuickHash == "" {
		srcHash, err := calculateQuickHash(src)
		if err != nil {
			return CompareResult{NeedsCopy: true, Reason: "try next strategy", Strategy: "quickhash"}, err
		}
		meta.QuickHash = srcHash
	}

	// Calculate destination quick hash
	dstHash, err := calculateQuickHash(dst)
	if err != nil {
		return CompareResult{NeedsCopy: true, Reason: "try next strategy", Strategy: "quickhash"}, err
	}

	// If quick hashes match, files are very likely identical
	if meta.QuickHash == dstHash {
		return CompareResult{NeedsCopy: false, Reason: "quick hash match", Strategy: "quickhash"}, nil
	}

	// Different quick hashes mean files are definitely different
	return CompareResult{NeedsCopy: true, Reason: "quick hash mismatch", Strategy: "quickhash"}, nil
}

func (q *QuickHashCompare) Priority() int {
	return 2
}

// FullChecksumCompare is most thorough but slowest
type FullChecksumCompare struct{}

func (f *FullChecksumCompare) Compare(src, dst string, meta *Metadata) (CompareResult, error) {
	// Calculate source checksum if not already done
	if meta.FullChecksum == "" {
		srcChecksum, err := calculateFullChecksum(src)
		if err != nil {
			return CompareResult{NeedsCopy: true, Reason: "checksum failed", Strategy: "checksum"}, err
		}
		meta.FullChecksum = srcChecksum
	}

	// Calculate destination checksum
	dstChecksum, err := calculateFullChecksum(dst)
	if err != nil {
		return CompareResult{NeedsCopy: true, Reason: "checksum failed", Strategy: "checksum"}, err
	}

	// Checksums match = files are identical
	if meta.FullChecksum == dstChecksum {
		return CompareResult{NeedsCopy: false, Reason: "checksum match", Strategy: "checksum"}, nil
	}

	// Different checksums = files are different
	return CompareResult{NeedsCopy: true, Reason: "checksum mismatch", Strategy: "checksum"}, nil
}

func (f *FullChecksumCompare) Priority() int {
	return 3
}

// Add the hash calculation functions
func calculateQuickHash(path string) (string, error) {
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

func calculateFullChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
