package sync

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RsyncOperation represents a type of rsync operation
type RsyncOperation int

const (
	AnalyzeChanges RsyncOperation = iota
	AnalyzeDeletions
	SyncFiles
	SyncDeletions
)

// RsyncOptions holds options for rsync commands
type RsyncOptions struct {
	Source          string
	Target          string
	IncludeDirs     []string
	ExcludePatterns []string
	LogFile         string
	DryRun          bool
	Delete          bool
}

// RsyncStats contains statistics from an rsync operation
type RsyncStats struct {
	FilesTransferred int
	TotalSize        string
	Duration         time.Duration
}

// buildBaseCommand constructs the base rsync command
func buildBaseCommand(opts RsyncOptions) []string {
	args := []string{"-a", "--stats"}

	// For dry-run mode, use less verbose output
	if opts.DryRun {
		// We don't want to see every file, just statistics
		args = append(args, "-n")
	} else {
		// For actual sync, show progress
		args = append(args, "--info=progress2")
	}

	// Add delete flag if specified
	if opts.Delete {
		args = append(args, "--delete")
	}

	// Add exclude patterns
	for _, pattern := range opts.ExcludePatterns {
		args = append(args, "--exclude="+pattern)
	}

	return args
}

// runForSingleDirectory executes rsync for a single directory
func runForSingleDirectory(dir, source, target string, baseArgs []string, logger *slog.Logger) (string, error) {
	sourceDir := source
	targetDir := target

	// If we're using include directories, append the directory to the paths
	if dir != "" {
		sourceDir = fmt.Sprintf("%s/%s/", source, dir)
		targetDir = fmt.Sprintf("%s/%s/", target, dir)
		
		// Ensure target directory exists
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			logger.Error("Failed to create target directory",
				"directory", targetDir,
				"error", err)
			return "", fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
		}
	}

	args := append([]string{}, baseArgs...)
	args = append(args, sourceDir, targetDir)

	cmd := exec.Command("rsync", args...)
	
	logger.Info("Running rsync command", 
		"command", fmt.Sprintf("rsync %s", strings.Join(args, " ")),
		"sourceDir", sourceDir,
		"targetDir", targetDir)
		
	fmt.Printf("Running rsync: %s to %s\n", sourceDir, targetDir)

	// Create buffers for stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	
	// Set up pipes for command output
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("error creating stdout pipe: %w", err)
	}
	
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("error creating stderr pipe: %w", err)
	}
	
	// Start the command
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting rsync: %w", err)
	}
	
	// Create a wait group to wait for both goroutines to finish
	var wg sync.WaitGroup
	wg.Add(2)
	
	// Process stdout
	go func() {
		defer wg.Done()
		
		scanner := bufio.NewScanner(stdoutPipe)
		progressRegex := regexp.MustCompile(`^\s*(\d+)%`)
		
		for scanner.Scan() {
			line := scanner.Text()
			stdoutBuf.WriteString(line + "\n")
			
			// Check if this is a progress line
			if matches := progressRegex.FindStringSubmatch(line); len(matches) > 1 {
				percent, _ := strconv.Atoi(matches[1])
				fmt.Printf("\rProgress: %d%%", percent)
				// Clear line if at 100%
				if percent >= 100 {
					fmt.Println()
				}
			} else if strings.HasPrefix(line, "Number of files:") ||
			          strings.HasPrefix(line, "Number of regular files transferred:") ||
			          strings.HasPrefix(line, "Total file size:") || 
			          strings.HasPrefix(line, "Total transferred file size:") {
				// Only print summary statistics
				fmt.Println(line)
			}
		}
	}()
	
	// Process stderr
	go func() {
		defer wg.Done()
		
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			stderrBuf.WriteString(line + "\n")
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", line)
		}
	}()
	
	// Wait for processing to complete
	wg.Wait()
	
	// Wait for the command to finish
	err = cmd.Wait()
	
	// Combine stdout and stderr
	output := stdoutBuf.String() + stderrBuf.String()
	
	// Log summary
	logger.Info("Rsync finished",
		"stdout_size", stdoutBuf.Len(),
		"stderr_size", stderrBuf.Len(),
		"error", err)
	
	return output, err
}

// parseRsyncStats extracts statistics from rsync output
func parseRsyncStats(output string) RsyncStats {
	var stats RsyncStats
	
	// Extract files transferred
	fileRegex := regexp.MustCompile(`Number of regular files transferred: (\d+)`)
	if matches := fileRegex.FindStringSubmatch(output); len(matches) > 1 {
		files, _ := strconv.Atoi(matches[1])
		stats.FilesTransferred = files
	}
	
	// Extract total size
	sizeRegex := regexp.MustCompile(`Total transferred file size: ([\d.,]+ [A-Za-z]+)`)
	if matches := sizeRegex.FindStringSubmatch(output); len(matches) > 1 {
		stats.TotalSize = matches[1]
	}
	
	return stats
}

// RunRsync executes an rsync operation
func RunRsync(op RsyncOperation, opts RsyncOptions, logger *slog.Logger) (RsyncStats, error) {
	startTime := time.Now()
	var stats RsyncStats
	
	// Set operation-specific flags
	switch op {
	case AnalyzeChanges:
		opts.DryRun = true
		opts.Delete = false
	case AnalyzeDeletions:
		opts.DryRun = true
		opts.Delete = true
	case SyncFiles:
		opts.DryRun = false
		opts.Delete = false
	case SyncDeletions:
		opts.DryRun = false
		opts.Delete = true
	}
	
	// Build the base command arguments after setting operation flags
	baseArgs := buildBaseCommand(opts)
	
	// Check if we're using include directories
	if len(opts.IncludeDirs) > 0 {
		// Run rsync for each included directory
		for _, dir := range opts.IncludeDirs {
			logger.Info("Processing directory", "directory", dir)
			fmt.Printf("\nProcessing directory: %s\n", dir)
			
			output, err := runForSingleDirectory(dir, opts.Source, opts.Target, baseArgs, logger)
			if err != nil {
				return stats, err
			}
			
			// Parse stats from this directory's output
			dirStats := parseRsyncStats(output)
			stats.FilesTransferred += dirStats.FilesTransferred
			// Note: Total size will be from the last directory processed
			if dirStats.TotalSize != "" {
				stats.TotalSize = dirStats.TotalSize
			}
		}
	} else {
		// Run rsync for the entire source/target
		output, err := runForSingleDirectory("", opts.Source, opts.Target, baseArgs, logger)
		if err != nil {
			return stats, err
		}
		
		stats = parseRsyncStats(output)
	}
	
	// Calculate duration
	stats.Duration = time.Since(startTime)
	
	return stats, nil
}

// CountDeletions counts the number of files that would be deleted
func CountDeletions(opts RsyncOptions, logger *slog.Logger) (int, []string, error) {
	count := 0
	var sampleDeletions []string
	
	// Ensure this is a dry run with delete flag
	opts.DryRun = true
	opts.Delete = true
	baseArgs := buildBaseCommand(opts)
	
	// Function to process a single directory
	processDeletions := func(dir, source, target string) (int, []string, error) {
		sourceDir := source
		targetDir := target
		
		if dir != "" {
			sourceDir = fmt.Sprintf("%s/%s/", source, dir)
			targetDir = fmt.Sprintf("%s/%s/", target, dir)
		}
		
		args := append([]string{}, baseArgs...)
		args = append(args, sourceDir, targetDir)
		
		cmd := exec.Command("rsync", args...)
		logger.Info("Running deletion analysis", 
			"command", fmt.Sprintf("rsync %s", strings.Join(args, " ")))
		
		fmt.Printf("Analyzing potential deletions: %s to %s\n", sourceDir, targetDir)
		
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		
		err := cmd.Run()
		if err != nil {
			logger.Error("Deletion analysis failed", 
				"stdout", stdout.String(),
				"stderr", stderr.String(),
				"error", err)
			return 0, nil, fmt.Errorf("error running deletion analysis: %w", err)
		}
		
		// Count deletions
		localCount := 0
		var localSamples []string
		
		scanner := bufio.NewScanner(&stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "deleting ") {
				filename := strings.TrimPrefix(line, "deleting ")
				// Don't print each deletion, just collect them
				localCount++
				if len(localSamples) < 5 {
					localSamples = append(localSamples, filename)
				}
			}
		}
		
		return localCount, localSamples, nil
	}
	
	// Check if we're using include directories
	if len(opts.IncludeDirs) > 0 {
		for _, dir := range opts.IncludeDirs {
			// Don't print this for each directory to reduce verbosity
		// fmt.Printf("\nAnalyzing deletions for directory: %s\n", dir)
			localCount, localSamples, err := processDeletions(dir, opts.Source, opts.Target)
			if err != nil {
				return count, sampleDeletions, err
			}
			
			count += localCount
			if len(sampleDeletions) < 5 {
				remainingSlots := 5 - len(sampleDeletions)
				if len(localSamples) < remainingSlots {
					sampleDeletions = append(sampleDeletions, localSamples...)
				} else {
					sampleDeletions = append(sampleDeletions, localSamples[:remainingSlots]...)
				}
			}
		}
	} else {
		localCount, localSamples, err := processDeletions("", opts.Source, opts.Target)
		if err != nil {
			return count, sampleDeletions, err
		}
		
		count = localCount
		sampleDeletions = localSamples
	}
	
	return count, sampleDeletions, nil
}