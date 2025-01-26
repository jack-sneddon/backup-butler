// internal/commands/root.go
package commands

import (
	"fmt"
	"os"

	"github.com/jack-sneddon/backup-butler/internal/commands/check"
	"github.com/jack-sneddon/backup-butler/internal/commands/sync" // Updated import
	"github.com/jack-sneddon/backup-butler/internal/commands/version"
	"github.com/jack-sneddon/backup-butler/internal/logger"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "backup-butler",
	Short: "A reliable media backup utility with data validation",
	Long: `Backup Butler is a command-line utility designed for reliable media backup 
    with high data integrity validation.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// 1. Setup flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")
	rootCmd.PersistentFlags().String("log-level", "error", "Log level (debug|info|warn|error)")

	// 2. Initialize logger in PreRun
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		level, _ := cmd.Flags().GetString("log-level")
		if err := logger.SetLevel(level); err != nil {
			return fmt.Errorf("invalid log level: %w", err)
		}
		return nil
	}

	// 3. Add commands
	rootCmd.AddCommand(version.NewVersionCmd())
	rootCmd.AddCommand(check.NewCheckCmd())
	rootCmd.AddCommand(sync.NewSyncCmd())
}
