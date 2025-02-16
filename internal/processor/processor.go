// internal/processor/processor.go
package processor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/logger"
	"go.uber.org/zap"
)

type directoryProcessor struct {
	opts *ProcessorOptions
	log  *zap.SugaredLogger
	sem  chan struct{} // Semaphore for thread limiting
}

func NewDirectoryProcessor(opts *ProcessorOptions) DirectoryProcessor {
	if opts == nil {
		opts = &ProcessorOptions{
			PreserveMetadata: true,
			BufferSize:       32768,
			MaxThreads:       4,
			StorageType:      "hdd",
		}
	}

	// Optimize settings based on storage type
	switch opts.StorageType {
	case "ssd":
		if opts.MaxThreads == 0 {
			opts.MaxThreads = 16
		}
		if opts.BufferSize == 0 {
			opts.BufferSize = 256 * 1024 // 256KB
		}
	case "network":
		if opts.MaxThreads == 0 {
			opts.MaxThreads = 8
		}
		if opts.BufferSize == 0 {
			opts.BufferSize = 1024 * 1024 // 1MB
		}
	default: // hdd
		if opts.MaxThreads == 0 {
			opts.MaxThreads = 4
		}
		if opts.BufferSize == 0 {
			opts.BufferSize = 32 * 1024 // 32KB
		}
	}

	return &directoryProcessor{
		opts: opts,
		log:  logger.Get(),
		sem:  make(chan struct{}, opts.MaxThreads),
	}
}

func (p *directoryProcessor) ProcessDirectory(sourcePath, targetPath string) error {
	// Ensure source directory exists
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		p.log.Errorw("Failed to access source directory",
			"path", sourcePath,
			"error", err)
		return fmt.Errorf("failed to access source directory: %w", err)
	}
	if !sourceInfo.IsDir() {
		p.log.Errorw("Source path is not a directory",
			"path", sourcePath)
		return fmt.Errorf("source path is not a directory: %s", sourcePath)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetPath, sourceInfo.Mode()); err != nil {
		p.log.Errorw("Failed to create target directory",
			"path", targetPath,
			"error", err)
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Read directory entries
	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		p.log.Errorw("Failed to read source directory",
			"path", sourcePath,
			"error", err)
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	if p.opts.Progress != nil {
		// Calculate directory totals
		var dirTotal int64
		var fileCount int
		for _, entry := range entries {
			if !entry.IsDir() {
				info, err := entry.Info()
				if err == nil {
					dirTotal += info.Size()
					fileCount++
				}
			}
		}

		if err := p.opts.Progress.StartDirectory(sourcePath, dirTotal, fileCount); err != nil {
			p.log.Warnw("Failed to start progress tracking", "error", err)
		}
	}

	// Process files concurrently with thread limiting
	var wg sync.WaitGroup
	errs := make(chan error, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			// Process subdirectory
			subSourcePath := filepath.Join(sourcePath, entry.Name())
			subTargetPath := filepath.Join(targetPath, entry.Name())

			if err := p.ProcessDirectory(subSourcePath, subTargetPath); err != nil {
				p.log.Errorw("Failed to process subdirectory",
					"directory", entry.Name(),
					"error", err)
				errs <- err
			}
			continue
		}

		wg.Add(1)
		go func(e os.DirEntry) {
			defer wg.Done()

			// Acquire semaphore
			p.sem <- struct{}{}
			defer func() { <-p.sem }()

			sourceFile := filepath.Join(sourcePath, e.Name())
			targetFile := filepath.Join(targetPath, e.Name())

			p.log.Debugw("Processing file",
				"source", sourceFile,
				"target", targetFile)

			if err := p.copyFile(sourceFile, targetFile); err != nil {
				p.log.Errorw("Failed to copy file",
					"file", e.Name(),
					"error", err)
				errs <- fmt.Errorf("failed to copy %s: %w", e.Name(), err)
			}

			// progress tracking
			if p.opts.Progress != nil {
				if info, err := e.Info(); err == nil {
					p.opts.Progress.UpdateProgress(info.Size())
				}
			}
		}(entry)
	}

	// Wait for all files to be processed
	wg.Wait()
	close(errs)

	// Check for any errors
	var processErrors []error
	for err := range errs {
		processErrors = append(processErrors, err)
	}

	// Finish directory progress tracking
	if p.opts.Progress != nil {
		if err := p.opts.Progress.FinishDirectory(); err != nil {
			p.log.Warnw("Failed to finish progress tracking", "error", err)
		}
	}

	if len(processErrors) > 0 {
		// Log all errors
		for _, err := range processErrors {
			p.log.Errorw("File processing error", "error", err)
		}
		return fmt.Errorf("failed to process %d files", len(processErrors))
	}

	return nil
}

func (p *directoryProcessor) copyFile(sourcePath, targetPath string) error {
	// Open source file
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	// Get source file info for metadata
	sourceInfo, err := source.Stat()
	if err != nil {
		return err
	}

	// Create target file
	target, err := os.OpenFile(targetPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return err
	}
	defer target.Close()

	// Ensure we have a valid buffer size
	bufferSize := p.opts.BufferSize
	if bufferSize <= 0 {
		bufferSize = 32768 // Default to 32KB if not set
	}

	// Copy content
	buf := make([]byte, bufferSize)
	if _, err := io.CopyBuffer(target, source, buf); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Preserve metadata if requested
	if p.opts.PreserveMetadata {
		modTime := sourceInfo.ModTime()
		if err := os.Chtimes(targetPath, time.Now(), modTime); err != nil {
			return fmt.Errorf("failed to preserve modification time: %w", err)
		}
	}

	return nil
}

func parseStorageType(t string) string {
	switch t {
	case "ssd", "SSD":
		return "ssd"
	case "network", "NETWORK":
		return "network"
	default:
		return "hdd"
	}
}
