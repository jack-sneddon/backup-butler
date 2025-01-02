// cmd/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sort"

	"github.com/jack-sneddon/backup-butler/internal/backup"
	"github.com/jack-sneddon/backup-butler/internal/config"
)

const (
	ExitSuccess = 0
	ExitError   = 1
)

// cmd/main.go
func main() {
	configPath := flag.String("config", "", "Path to configuration file")
	validateOnly := flag.Bool("validate", false, "Validate configuration only")
	featureTest := flag.Bool("feature-test", false, "Run comprehensive feature tests")
	showVersions := flag.Bool("versions", false, "Show backup version history")
	verifyIntegrity := flag.Bool("verify", false, "Verify integrity of backed up files")
	showIssues := flag.Bool("show-issues", false, "Show detected integrity issues")

	flag.Parse()

	if *configPath == "" {
		fmt.Println("Error: -config flag is required")
		flag.Usage()
		os.Exit(ExitError)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		os.Exit(ExitError)
	}

	// Add version display before other operations
	if *showVersions {
		if err := displayVersionHistory(cfg); err != nil {
			fmt.Printf("Failed to show version history: %v\n", err)
			os.Exit(ExitError)
		}
		os.Exit(ExitSuccess)
	}

	fmt.Println("Configuration is valid!")
	fmt.Printf("Source: %s\n", cfg.SourceDirectory)
	fmt.Printf("Target: %s\n", cfg.TargetDirectory)
	fmt.Printf("Folders: %v\n", cfg.FoldersToBackup)

	if *featureTest {
		fmt.Println("\nNot yet implemented")
		/*
			if err := runFeatureTests(cfg); err != nil {
				fmt.Printf("Feature tests failed: %v\n", err)
				os.Exit(ExitError)
			}
			fmt.Println("\nFeature tests completed successfully!")
		*/
		os.Exit(ExitSuccess)
	}

	if *validateOnly {
		os.Exit(ExitSuccess)
	}

	if *verifyIntegrity {
		if err := verifyBackupIntegrity(cfg); err != nil {
			fmt.Printf("Integrity verification failed: %v\n", err)
			os.Exit(ExitError)
		}
		os.Exit(ExitSuccess)
	}

	if *showIssues {
		if err := showIntegrityIssues(cfg); err != nil {
			fmt.Printf("Failed to show integrity issues: %v\n", err)
			os.Exit(ExitError)
		}
		os.Exit(ExitSuccess)
	}

	if err := runBackup(cfg); err != nil {
		fmt.Printf("Backup failed: %v\n", err)
		os.Exit(ExitError)
	}
}

// cmd/main.go
func displayVersionHistory(cfg *config.Config) error {
	service, err := backup.NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}

	versions, err := service.GetVersionHistory()
	if err != nil {
		return fmt.Errorf("failed to get version history: %w", err)
	}

	fmt.Println("\nBackup Version History:")
	fmt.Println("======================")

	for _, v := range versions {
		fmt.Printf("\nVersion: %s\n", v.ID)
		fmt.Printf("Started: %v\n", v.StartTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("Duration: %v\n", v.EndTime.Sub(v.StartTime))

		// Overall statistics
		fmt.Printf("\nOverall Statistics:\n")
		fmt.Printf("  Total Files: %d (%.2f MB)\n",
			v.Stats.TotalFiles,
			float64(v.Stats.BytesCopied+v.Stats.BytesSkipped)/(1024*1024))
		fmt.Printf("  Copied: %d files (%.2f MB)\n",
			v.Stats.FilesCopied,
			float64(v.Stats.BytesCopied)/(1024*1024))
		fmt.Printf("  Skipped: %d files (%.2f MB)\n",
			v.Stats.FilesSkipped,
			float64(v.Stats.BytesSkipped)/(1024*1024))
		if v.Stats.FilesFailed > 0 {
			fmt.Printf("  Failed: %d files\n", v.Stats.FilesFailed)
		}

		// Directory statistics
		if len(v.DirStats) > 0 {
			fmt.Printf("\nDirectory Statistics:\n")
			// Sort directories for consistent display
			dirs := make([]string, 0, len(v.DirStats))
			for dir := range v.DirStats {
				dirs = append(dirs, dir)
			}
			sort.Strings(dirs)

			for _, dir := range dirs {
				stats := v.DirStats[dir]
				fmt.Printf("  %s:\n", dir)
				fmt.Printf("    Files: %d total (%.2f MB)\n",
					stats.TotalFiles,
					float64(stats.TotalBytes)/(1024*1024))
				fmt.Printf("    Copied: %d files (%.2f MB)\n",
					stats.CopiedFiles,
					float64(stats.CopiedBytes)/(1024*1024))
				fmt.Printf("    Skipped: %d files (%.2f MB)\n",
					stats.SkippedFiles,
					float64(stats.SkippedBytes)/(1024*1024))
				if stats.FailedFiles > 0 {
					fmt.Printf("    Failed: %d files\n", stats.FailedFiles)
				}
			}
		}
		fmt.Println("----------------------")
	}

	return nil
}

/*
func runFeatureTests(cfg *config.Config) error {
	featureTest, err := test.NewFeatureTest(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize feature tests: %w", err)
	}
	return featureTest.RunFeatureTests()
}
*/

func runBackup(cfg *config.Config) error {
	fmt.Println("\nInitializing backup service...")

	service, err := backup.NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize backup service: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal. Canceling backup...")
		cancel()
	}()

	fmt.Println("\nStarting backup operation...")
	return service.Backup(ctx)
}

func verifyBackupIntegrity(cfg *config.Config) error {
	service, err := backup.NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}

	fmt.Println("Verifying backup integrity...")

	// Use the service to verify integrity
	issues, err := service.GetIntegrityIssues()
	if err != nil {
		return fmt.Errorf("failed to retrieve integrity issues: %w", err)
	}

	if len(issues) > 0 {
		fmt.Printf("Found %d integrity issues. Use --show-issues for details.\n", len(issues))
	} else {
		fmt.Println("No integrity issues found.")
	}

	return nil
}

func showIntegrityIssues(cfg *config.Config) error {
	service, err := backup.NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}

	issues, err := service.GetIntegrityIssues()
	if err != nil {
		return fmt.Errorf("failed to get integrity issues: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No integrity issues found.")
		return nil
	}

	fmt.Printf("\nFound %d integrity issues:\n", len(issues))
	for _, issue := range issues {
		fmt.Printf("\nFile: %s\n", issue.Path)
		fmt.Printf("Severity: %s\n", issue.Severity)
		fmt.Printf("Issues:\n")
		for _, detail := range issue.Issues {
			fmt.Printf("  - %s\n", detail)
		}
		if issue.Details != "" {
			fmt.Printf("Details: %s\n", issue.Details)
		}
		fmt.Println("---")
	}

	return nil
}
