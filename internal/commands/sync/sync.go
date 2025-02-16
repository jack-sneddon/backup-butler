// internal/commands/sync/sync.go
package sync

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jack-sneddon/backup-butler/internal/config"
	"github.com/jack-sneddon/backup-butler/internal/logger"
	"github.com/jack-sneddon/backup-butler/internal/processor"
	"github.com/jack-sneddon/backup-butler/internal/progress"
	"github.com/spf13/cobra"
)

func NewSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize source to target using configuration",
		RunE:  runSync,
	}

	// Required flags
	cmd.Flags().StringSliceP("folders", "d", []string{}, "specific folders to sync")

	// TODO: Implement these flags in Phase 4
	cmd.Flags().BoolP("resume", "r", false, "resume from last checkpoint")
	cmd.Flags().BoolP("force", "f", false, "override safety checks")
	cmd.Flags().BoolP("quiet", "q", false, "suppress configuration output") // Add this

	return cmd
}

func runSync(cmd *cobra.Command, args []string) error {
	log := logger.Get()
	cfgFile := cmd.Root().PersistentFlags().Lookup("config").Value.String()

	log.Debugw("Starting sync command execution")
	log.Infow("Loading configuration", "file", cfgFile)

	cfg, err := loadAndValidateConfig(cmd, cfgFile)
	if err != nil {
		return err
	}

	displayConfiguration(cmd, cfg)

	return processDirectories(cfg)
}

func loadAndValidateConfig(cmd *cobra.Command, cfgFile string) (*config.Config, error) {
	log := logger.Get()

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		msg := strings.TrimPrefix(err.Error(), "invalid configuration: ")
		msg = strings.TrimPrefix(msg, "failed to load config: ")
		log.Error(msg)
		return nil, fmt.Errorf("%s", msg)
	}

	if cmd.Flag("log-level").Value.String() == "debug" {
		log.Debugw("Configuration loaded successfully",
			"source", cfg.Source,
			"target", cfg.Target)
	}

	// Handle folder overrides
	if folders, _ := cmd.Flags().GetStringSlice("folders"); len(folders) > 0 {
		log.Infow("Using folder override", "folders", folders)
		cfg.Folders = folders
	}

	log.Infow("Starting sync",
		"source", cfg.Source,
		"target", cfg.Target,
		"folders", cfg.Folders,
		"source deviceType", cfg.Storage.Source.Type,
		"target deviceType", cfg.Storage.Target.Type)

	return cfg, nil
}

func displayConfiguration(cmd *cobra.Command, cfg *config.Config) {
	// Skip display if quiet flag is set
	if quiet, _ := cmd.Flags().GetBool("quiet"); quiet {
		return
	}

	// Only display if not at warn/error level
	if cmd.Flag("log-level").Value.String() == "warn" ||
		cmd.Flag("log-level").Value.String() == "error" {
		return
	}

	fmt.Println("\nConfiguration:")
	fmt.Printf("├── Locations\n")
	fmt.Printf("│   ├── Source: %s\n", cfg.Source)
	fmt.Printf("│   └── Target: %s\n", cfg.Target)

	if len(cfg.Folders) > 0 {
		fmt.Printf("├── Folders\n")
		for i, folder := range cfg.Folders {
			if i == len(cfg.Folders)-1 {
				fmt.Printf("│   └── %s\n", folder)
			} else {
				fmt.Printf("│   ├── %s\n", folder)
			}
		}
	}

	fmt.Printf("├── Storage\n")
	fmt.Printf("│   ├── Source Type: %s\n", cfg.Storage.Source.Type)
	fmt.Printf("│   └── Target Type: %s\n", cfg.Storage.Target.Type)

	fmt.Printf("└── Validation\n")
	fmt.Printf("    ├── Algorithm: %s\n", cfg.Comparison.Algorithm)
	fmt.Printf("    └── Level: %s\n", cfg.Comparison.Level)
}

func processDirectories(cfg *config.Config) error {
	// Create progress tracker
	tracker := progress.NewTracker()
	if err := tracker.Start(); err != nil {
		return fmt.Errorf("failed to start progress tracking: %w", err)
	}
	defer tracker.Stop() // Will now wait properly for display

	opts := &processor.ProcessorOptions{
		PreserveMetadata: true,
		BufferSize:       cfg.Comparison.BufferSize,
		MaxThreads:       cfg.Storage.Source.MaxThreads,
		StorageType:      cfg.Storage.Source.Type,
		Progress:         tracker,
	}
	proc := processor.NewDirectoryProcessor(opts)

	if len(cfg.Folders) > 0 {
		return processFolders(cfg, proc)
	}
	return processRootDirectory(cfg, proc)

}

func processFolders(cfg *config.Config, proc processor.DirectoryProcessor) error {
	log := logger.Get()

	for _, folder := range cfg.Folders {
		sourcePath := filepath.Join(cfg.Source, folder)
		targetPath := filepath.Join(cfg.Target, folder)

		log.Infow("Processing folder",
			"folder", folder,
			"source", sourcePath,
			"target", targetPath)

		if err := proc.ProcessDirectory(sourcePath, targetPath); err != nil {
			log.Errorw("Failed to process folder",
				"folder", folder,
				"error", err)
			return err
		}
	}
	return nil
}

func processRootDirectory(cfg *config.Config, proc processor.DirectoryProcessor) error {
	log := logger.Get()

	if err := proc.ProcessDirectory(cfg.Source, cfg.Target); err != nil {
		log.Errorw("Failed to process root directory", "error", err)
		return err
	}
	return nil
}
