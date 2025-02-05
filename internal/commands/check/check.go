// internal/commands/check/check.go
package check

import (
	"fmt"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/config"
	"github.com/jack-sneddon/backup-butler/internal/logger"
	"github.com/jack-sneddon/backup-butler/internal/scan"
	"github.com/spf13/cobra"
)

// ValidationLevel defines how thorough the check should be
type ValidationLevel string

const (
	Quick    ValidationLevel = "quick"    // Size and modification time only
	Standard ValidationLevel = "standard" // Includes basic hash comparison
	Deep     ValidationLevel = "deep"     // Full content verification
)

func NewCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [source] [target]",
		Short: "Check backup integrity",
		Long: `Check performs integrity validation between source and target directories.
Validation levels:
  quick:     Compare size and modification times only
  standard:  Include hash comparison (default)
  deep:      Perform full content verification`,
		RunE: runCheck,
	}

	cmd.Flags().StringP("level", "l", "standard", "validation level (quick|standard|deep)")
	cmd.Flags().StringP("output", "o", "text", "output format (text|csv|html)")

	return cmd
}

func runCheck(cmd *cobra.Command, args []string) error {
	log := logger.Get()

	cfgFile := cmd.Root().PersistentFlags().Lookup("config").Value.String()
	log.Debugw("Loading config", "file", cfgFile)

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	if cfg.Validation != nil {
		log.Debugw("check.go::runCheck - Config after loading",
			"validation", cfg.Validation)
	} else {
		log.Debugw("check.go::runCheck - No validation config found")
	}

	log.Debugw("Config loaded",
		"source", cfg.Source,
		"target", cfg.Target,
		"excludePatterns", cfg.Exclude,
		"includeFolders", cfg.Folders,
		"logging", cfg.Logging)

	// Get validation level (prefer config unless overridden by command line)
	var level string

	// 1. Get default from config if available
	// 1. Get default from config if available
	if cfg.Validation != nil && cfg.Validation.DefaultLevel != "" {
		level = cfg.Validation.DefaultLevel
		log.Debugw("Using config validation level", "level", level)
	}

	// 1. Get default from config if available
	if cfg.Validation != nil && cfg.Validation.DefaultLevel != "" {
		level = cfg.Validation.DefaultLevel
		log.Debugw("Using config validation level", "level", level)
		log.Debugw("check.go::runCheck() - Using config validation level",
			"level", level,
			"configLevel", cfg.Validation.DefaultLevel)

	} else {
		log.Debugw("check.go::runCheck() - No validation level in config")
	}

	// 2. Check command line flag
	// 2. Check command line flag
	cmdLevel, err := cmd.Flags().GetString("level")
	if err != nil {
		return err
	}
	if cmd.Flag("level").Changed { // Only override if explicitly set
		level = cmdLevel
	}

	// 3. Set standard as default if neither is specified
	if level == "" {
		level = string(Standard)
		log.Debugw("check.go::runCheck() - Using default validation level",
			"level", level)
	}

	// 4. Validate the final level
	if !isValidLevel(level) {
		return fmt.Errorf("invalid validation level: %s", level)
	}

	log.Debugw("Final validation level set",
		"level", level,
		"hasValidation", cfg.Validation != nil,
		"hasDefaultLevel", cfg.Validation != nil && cfg.Validation.DefaultLevel != "")

	if cfg.Validation != nil {
		log.Debugw("Validation config",
			"default", cfg.Validation.DefaultLevel)
	}

	// Create scanner options from config
	opts := &scan.ScannerOptions{
		ExcludePatterns:  cfg.Exclude,
		IncludeFolders:   cfg.Folders,
		BufferSize:       cfg.Comparison.BufferSize,
		MaxDepth:         -1,
		DefaultLevel:     level,
		ValidationConfig: cfg.Validation,
	}

	log.Debugw("check.go:runCheck() - Scanner options created",
		"validationConfig", opts.ValidationConfig)

	scanner := scan.NewScanner(opts)

	// Start progress display
	doneChan := make(chan bool)
	go displayProgress(scanner.GetProgress(), doneChan)

	// Perform the scan
	progress, err := scanner.Scan(cfg.Source)
	if err != nil {
		doneChan <- true
		return err
	}

	// Print final summary
	doneChan <- true
	printSummary(cfg, progress)

	// Perform comparison based on validation level
	comparisons, err := scanner.Compare(cfg.Source, cfg.Target)
	if err != nil {
		return err
	}

	// Print comparison results
	printResults(comparisons)

	return nil
}

func displayProgress(progress *scan.Progress, done chan bool) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			fmt.Print("\033[2K\r") // Clear the line
			return
		case <-ticker.C:
			if progress.TotalBytes > 0 {
				percentage := float64(progress.ProcessedBytes) / float64(progress.TotalBytes) * 100
				fmt.Printf("\033[2K\r%s - %.1f%% (%d/%d files, %s/%s)",
					progress.CurrentDir,
					percentage,
					progress.ScannedFiles,
					progress.TotalFiles,
					formatBytes(progress.ProcessedBytes),
					formatBytes(progress.TotalBytes))
			}
		}
	}
}

func printSummary(cfg *config.Config, progress *scan.Progress) {
	fmt.Printf("\nScan Results:\n")
	fmt.Printf("├── Locations\n")
	fmt.Printf("│   ├── Source: %s\n", cfg.Source)
	fmt.Printf("│   └── Target: %s\n", cfg.Target)

	fmt.Printf("├── Summary\n")
	fmt.Printf("│   ├── Directories: %d\n", progress.ScannedDirs)
	fmt.Printf("│   ├── Files: %d\n", progress.ScannedFiles)
	fmt.Printf("│   ├── Total Size: %s\n", formatBytes(progress.TotalBytes))
	if progress.ExcludedFiles > 0 || progress.ExcludedDirs > 0 {
		fmt.Printf("│   ├── Excluded Files: %d\n", progress.ExcludedFiles)
		fmt.Printf("│   └── Excluded Directories: %d\n", progress.ExcludedDirs)
	} else {
		fmt.Printf("│   └── No Exclusions\n")
	}

	if len(progress.Errors) > 0 {
		fmt.Printf("├── Scan Errors\n")
		for i, err := range progress.Errors {
			if i == len(progress.Errors)-1 {
				fmt.Printf("│   └── %s\n", err)
			} else {
				fmt.Printf("│   ├── %s\n", err)
			}
		}
	}
}

func printResults(comparisons []*scan.FileComparison) {
	var matches, new, missing, differs, errors int

	fmt.Printf("└── File Status\n")
	for _, comp := range comparisons {
		switch comp.Status {
		case scan.StatusMatch:
			matches++
		case scan.StatusNew:
			new++
		case scan.StatusMissing:
			missing++
		case scan.StatusDiffer:
			differs++
		case scan.StatusError:
			errors++
		}
		// Add validation level to output
		levelStr := ""
		if comp.Level != "" {
			levelStr = fmt.Sprintf(" [%s]", comp.Level)
		}
		fmt.Printf("    %c %s%s\n", comp.Status, comp.Path, levelStr)
	}

	// Print statistics
	fmt.Printf("\nResults Summary:\n")
	fmt.Printf("├── Matched:  %d files\n", matches)
	fmt.Printf("├── New:      %d files\n", new)
	fmt.Printf("├── Missing:  %d files\n", missing)
	fmt.Printf("├── Modified: %d files\n", differs)
	fmt.Printf("└── Errors:   %d files\n", errors)
}

func isValidLevel(level string) bool {
	switch ValidationLevel(level) {
	case Quick, Standard, Deep:
		return true
	}
	return false
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
