// internal/commands/sync/sync.go
package sync

import (
	"fmt"

	"github.com/jack-sneddon/backup-butler/internal/config"
	"github.com/spf13/cobra"
)

func runSync(cmd *cobra.Command, args []string) error {
	cfgFile := cmd.Root().PersistentFlags().Lookup("config").Value.String()

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get folder overrides if specified
	if folders, _ := cmd.Flags().GetStringSlice("folders"); len(folders) > 0 {
		cfg.Folders = folders
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
	fmt.Printf("│   ├── Device: %s\n", cfg.Storage.DeviceType)
	fmt.Printf("│   └── Threads: %d\n", cfg.Storage.MaxThreads)

	fmt.Printf("└── Validation\n")
	fmt.Printf("    ├── Algorithm: %s\n", cfg.Comparison.Algorithm)
	fmt.Printf("    └── Level: %s\n", cfg.Comparison.Level)

	return nil
}

// internal/commands/sync/sync.go
func NewSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize source to target using configuration",
		RunE:  runSync,
	}

	cmd.Flags().StringSliceP("folders", "d", []string{}, "specific folders to sync")
	cmd.Flags().BoolP("resume", "r", false, "resume from last checkpoint")
	cmd.Flags().BoolP("force", "f", false, "override safety checks")

	return cmd
}
