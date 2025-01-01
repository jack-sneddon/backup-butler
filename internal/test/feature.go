// internal/test/feature_test.go
package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/config"
	"github.com/jack-sneddon/backup-butler/internal/storage"
	"github.com/jack-sneddon/backup-butler/internal/types"
	"github.com/jack-sneddon/backup-butler/internal/version"
)

type FeatureTest struct {
	config         *config.Config
	storageManager *storage.Manager
	copier         *storage.Copier
	versionMgr     *version.Manager
}

func NewFeatureTest(cfg *config.Config) (*FeatureTest, error) {
	storageManager := storage.NewManager(cfg.BufferSize)
	copier := storage.NewCopier(storageManager, cfg.BufferSize)
	versionMgr, err := version.NewManager(cfg.TargetDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize version manager: %w", err)
	}

	return &FeatureTest{
		config:         cfg,
		storageManager: storageManager,
		copier:         copier,
		versionMgr:     versionMgr,
	}, nil
}

// RunFeatureTests runs all feature tests
func (t *FeatureTest) RunFeatureTests() error {
	fmt.Println("\nRunning Backup Butler Feature Tests...")

	// First backup - all files should be copied
	fmt.Println("\n=== Test 1: Initial Backup ===")
	if err := t.runInitialBackup(); err != nil {
		return fmt.Errorf("initial backup test failed: %w", err)
	}

	// Second backup - files should be skipped
	fmt.Println("\n=== Test 2: Incremental Backup (No Changes) ===")
	if err := t.runIncrementalBackup(); err != nil {
		return fmt.Errorf("incremental backup test failed: %w", err)
	}

	// Modified file backup
	fmt.Println("\n=== Test 3: Modified File Backup ===")
	if err := t.runModifiedBackup(); err != nil {
		return fmt.Errorf("modified backup test failed: %w", err)
	}

	// Display version history
	if err := t.displayVersionHistory(); err != nil {
		return fmt.Errorf("failed to display version history: %w", err)
	}

	return nil
}

// internal/test/feature.go

func (t *FeatureTest) displayVersionHistory() error {
	versions, err := t.versionMgr.GetVersions()
	if err != nil {
		return fmt.Errorf("failed to get versions: %w", err)
	}

	fmt.Printf("\nBackup History:\n")
	fmt.Printf("---------------\n")
	for _, v := range versions {
		fmt.Printf("Version: %s\n", v.ID)
		fmt.Printf("Status: %s\n", v.Status)
		fmt.Printf("Duration: %v\n", v.EndTime.Sub(v.StartTime))
		fmt.Printf("Files: %d total (%s)\n",
			v.TotalFiles,
			version.FormatSize(v.TotalBytes))
		fmt.Printf("Copied: %d files (%s)\n",
			v.CopiedFiles,
			version.FormatSize(v.CopiedBytes))
		fmt.Printf("Skipped: %d files (%s)\n",
			v.SkippedFiles,
			version.FormatSize(v.SkippedBytes))
		if v.FailedFiles > 0 {
			fmt.Printf("Failed: %d files\n", v.FailedFiles)
		}
		fmt.Printf("---------------\n")
	}
	return nil
}

func (t *FeatureTest) runInitialBackup() error {
	// Start new version
	currentVersion := t.versionMgr.StartNewVersion(t.config)
	fmt.Printf("Started version: %s\n", currentVersion.ID)

	if err := t.testSmallFile("initial"); err != nil {
		t.versionMgr.CompleteVersion("failed")
		return err
	}

	if err := t.testLargeFile(); err != nil {
		t.versionMgr.CompleteVersion("failed")
		return err
	}

	return t.versionMgr.CompleteVersion("completed")
}

func (t *FeatureTest) runIncrementalBackup() error {
	// Wait a moment to ensure different timestamp
	time.Sleep(time.Second)

	currentVersion := t.versionMgr.StartNewVersion(t.config)
	fmt.Printf("Started version: %s\n", currentVersion.ID)

	// Run same tests - files should be skipped
	if err := t.testSmallFile("incremental"); err != nil {
		t.versionMgr.CompleteVersion("failed")
		return err
	}

	if err := t.testLargeFile(); err != nil {
		t.versionMgr.CompleteVersion("failed")
		return err
	}

	return t.versionMgr.CompleteVersion("completed")
}

func (t *FeatureTest) runModifiedBackup() error {
	// Wait a moment to ensure different timestamp
	time.Sleep(time.Second)

	// Modify test file
	if err := t.testSmallFile("modified"); err != nil {
		t.versionMgr.CompleteVersion("failed")
		return err
	}

	if err := t.testLargeFile(); err != nil {
		t.versionMgr.CompleteVersion("failed")
		return err
	}

	return t.versionMgr.CompleteVersion("completed")
}

func (t *FeatureTest) testSmallFile(phase string) error {
	// Create test directories if they don't exist
	testSourceDir := filepath.Join(t.config.SourceDirectory, t.config.FoldersToBackup[0])
	if err := os.MkdirAll(testSourceDir, 0755); err != nil {
		return fmt.Errorf("failed to create test source directory: %w", err)
	}

	// Create or modify test file based on phase
	testFile := filepath.Join(testSourceDir, "test.txt")
	var content string
	switch phase {
	case "initial":
		content = "This is a test file for backup-butler storage operations.\n"
	case "incremental":
		// Don't modify the file
		content = "This is a test file for backup-butler storage operations.\n"
	case "modified":
		content = "This is a modified test file for backup-butler storage operations.\n"
	}

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create/modify test file: %w", err)
	}
	fmt.Printf("Created/Modified test file: %s\n", testFile)

	// Get and display metadata
	meta, err := t.storageManager.GetMetadata(testFile)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}
	t.displayMetadata(meta)

	// Calculate relative path for version history
	relPath, err := filepath.Rel(t.config.SourceDirectory, testFile)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	// Test file copy
	targetFile := filepath.Join(t.config.TargetDirectory, relPath)

	// Compare files first
	compareResult, err := t.storageManager.Compare(testFile, targetFile, t.versionMgr)
	if err != nil {
		compareResult = storage.CompareResult{NeedsCopy: true, Reason: "comparison failed"}
	}

	if compareResult.NeedsCopy {
		fmt.Printf("\nCopying to: %s\n", targetFile)
		result, err := t.copier.Copy(context.Background(), testFile, targetFile)
		if err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
		// fmt.Printf("Copied %d bytes in %v\n", result.BytesCopied, result.Duration)
		fmt.Printf("Copied %s in %v\n",
			version.FormatSize(result.BytesCopied),
			result.Duration)

		// Verify copy
		if err := t.copier.VerifyCopy(testFile, targetFile); err != nil {
			return fmt.Errorf("copy verification failed: %w", err)
		}
		fmt.Println("Copy verification successful!")

		// Record as copied
		t.versionMgr.RecordFile(relPath, version.FileResult{
			Path:         meta.Path,
			Size:         meta.Size,
			ModTime:      meta.ModTime,
			Checksum:     meta.Checksum,
			Status:       "copied",
			CopyDuration: result.Duration,
			Metadata:     meta,
		})
	} else {
		fmt.Printf("\nFile unchanged, skipping: %s\n", targetFile)
		// Record as skipped
		t.versionMgr.RecordFile(relPath, version.FileResult{
			Path:     meta.Path,
			Size:     meta.Size,
			ModTime:  meta.ModTime,
			Checksum: meta.Checksum,
			Status:   "skipped",
			Metadata: meta,
		})
	}

	fmt.Printf("\nComparison result:\n")
	fmt.Printf("Needs copy: %v\n", compareResult.NeedsCopy)
	fmt.Printf("Reason: %s\n", compareResult.Reason)

	return nil
}

func (t *FeatureTest) testLargeFile() error {
	var largeFile string
	var largeFileSize int64

	sourceDir := filepath.Join(t.config.SourceDirectory, t.config.FoldersToBackup[0])
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Size() > 1024*1024 {
			largeFile = path
			largeFileSize = info.Size()
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan for large file: %w", err)
	}

	if largeFile == "" {
		fmt.Println("\nNo large files found for testing")
		return nil
	}

	fmt.Printf("\nTesting with large file: %s (%s)\n",
		filepath.Base(largeFile),
		version.FormatSize(largeFileSize))

	meta, err := t.storageManager.GetMetadata(largeFile)
	if err != nil {
		return fmt.Errorf("failed to get large file metadata: %w", err)
	}
	t.displayMetadata(meta)

	relPath, err := filepath.Rel(t.config.SourceDirectory, largeFile)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	targetFile := filepath.Join(t.config.TargetDirectory, relPath)

	compareResult, err := t.storageManager.Compare(largeFile, targetFile, t.versionMgr)
	// ... rest remains
	if err != nil {
		compareResult = storage.CompareResult{NeedsCopy: true, Reason: "comparison failed"}
	}

	if compareResult.NeedsCopy {
		fmt.Printf("\nCopying to: %s\n", targetFile)
		result, err := t.copier.Copy(context.Background(), largeFile, targetFile)
		if err != nil {
			return fmt.Errorf("failed to copy large file: %w", err)
		}
		fmt.Printf("Copied %s in %v (%.2f MB/s)\n",
			version.FormatSize(result.BytesCopied),
			result.Duration,
			float64(result.BytesCopied)/float64(result.Duration.Seconds())/(1024*1024))

		fmt.Println("Verifying copy...")
		if err := t.copier.VerifyCopy(largeFile, targetFile); err != nil {
			return fmt.Errorf("large file copy verification failed: %w", err)
		}
		fmt.Println("Large file copy verification successful!")

		t.versionMgr.RecordFile(relPath, version.FileResult{
			Path:         meta.Path,
			Size:         meta.Size,
			ModTime:      meta.ModTime,
			Checksum:     meta.Checksum,
			Status:       "copied",
			CopyDuration: result.Duration,
			Metadata:     meta,
		})
	} else {
		fmt.Printf("File unchanged, skipping: %s\n", targetFile)
		t.versionMgr.RecordFile(relPath, version.FileResult{
			Path:     meta.Path,
			Size:     meta.Size,
			ModTime:  meta.ModTime,
			Checksum: meta.Checksum,
			Status:   "skipped",
			Metadata: meta,
		})
	}

	fmt.Printf("\nComparison result:\n")
	fmt.Printf("Needs copy: %v\n", compareResult.NeedsCopy)
	fmt.Printf("Reason: %s\n", compareResult.Reason)

	return nil
}

func (t *FeatureTest) displayMetadata(meta types.FileMetadata) {
	fmt.Printf("\nFile metadata:\n")
	fmt.Printf("Size: %s\n", version.FormatSize(meta.Size))
	fmt.Printf("Modified: %v\n", meta.ModTime)
	fmt.Printf("Checksum: %s\n", meta.Checksum)
}
