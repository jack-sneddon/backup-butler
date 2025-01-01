// internal/storage/checksum.go
package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// NewChecksumCalculator creates a new checksum calculator
func NewChecksumCalculator() *ChecksumCalculator {
	return &ChecksumCalculator{}
}

// CalculateChecksum computes SHA-256 checksum of a file
func (c *ChecksumCalculator) CalculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
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

// VerifyChecksum calculates and compares checksum with expected value
func (c *ChecksumCalculator) VerifyChecksum(path, expectedChecksum string) error {
	actualChecksum, err := c.CalculateChecksum(path)
	if err != nil {
		return err
	}

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s",
			expectedChecksum, actualChecksum)
	}

	return nil
}
