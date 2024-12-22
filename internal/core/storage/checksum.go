package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

type ChecksumCalculator struct{}

func NewChecksumCalculator() *ChecksumCalculator {
	return &ChecksumCalculator{}
}

func (c *ChecksumCalculator) CalculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
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
