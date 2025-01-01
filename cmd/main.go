// cmd/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/jack-sneddon/backup-butler/internal/backup"
	"github.com/jack-sneddon/backup-butler/internal/config"
	"github.com/jack-sneddon/backup-butler/internal/test"
)

const (
	ExitSuccess = 0
	ExitError   = 1
)

func main() {
	configPath := flag.String("config", "", "Path to configuration file")
	validateOnly := flag.Bool("validate", false, "Validate configuration only")
	featureTest := flag.Bool("feature-test", false, "Run comprehensive feature tests")
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

	fmt.Println("Configuration is valid!")
	fmt.Printf("Source: %s\n", cfg.SourceDirectory)
	fmt.Printf("Target: %s\n", cfg.TargetDirectory)
	fmt.Printf("Folders: %v\n", cfg.FoldersToBackup)

	if *featureTest {
		if err := runFeatureTests(cfg); err != nil {
			fmt.Printf("Feature tests failed: %v\n", err)
			os.Exit(ExitError)
		}
		fmt.Println("\nFeature tests completed successfully!")
		os.Exit(ExitSuccess)
	}

	if *validateOnly {
		os.Exit(ExitSuccess)
	}

	if err := runBackup(cfg); err != nil {
		fmt.Printf("Backup failed: %v\n", err)
		os.Exit(ExitError)
	}
}

func runFeatureTests(cfg *config.Config) error {
	featureTest, err := test.NewFeatureTest(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize feature tests: %w", err)
	}
	return featureTest.RunFeatureTests()
}

func runBackup(cfg *config.Config) error {
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
