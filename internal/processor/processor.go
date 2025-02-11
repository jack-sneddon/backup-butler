// internal/processor/processor.go
package processor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/logger"
	"go.uber.org/zap"
)

type directoryProcessor struct {
	opts *ProcessorOptions
	log  *zap.SugaredLogger
}

func NewDirectoryProcessor(opts *ProcessorOptions) DirectoryProcessor {
	if opts == nil {
		opts = &ProcessorOptions{
			PreserveMetadata: true,
			BufferSize:       32768, // 32KB default
		}
	}
	return &directoryProcessor{
		opts: opts,
		log:  logger.Get(),
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

	p.log.Debugw("Processing directory",
		"source", sourcePath,
		"target", targetPath,
		"fileCount", len(entries))

	// Process each file in the directory
	for _, entry := range entries {
		if entry.IsDir() {
			p.log.Debugw("Skipping subdirectory", "name", entry.Name())
			continue // Skip subdirectories for now
		}

		sourceFile := filepath.Join(sourcePath, entry.Name())
		targetFile := filepath.Join(targetPath, entry.Name())

		p.log.Debugw("Copying file",
			"source", sourceFile,
			"target", targetFile)

		if err := p.copyFile(sourceFile, targetFile); err != nil {
			p.log.Errorw("Failed to copy file",
				"file", entry.Name(),
				"error", err)
			return fmt.Errorf("failed to copy %s: %w", entry.Name(), err)
		}
	}

	p.log.Infow("Directory processing complete",
		"source", sourcePath,
		"target", targetPath)

	return nil
}

// internal/processor/processor.go

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
