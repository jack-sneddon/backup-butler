// internal/commands/root.go
package commands

import (
	"os"

	"github.com/jack-sneddon/backup-butler/internal/commands/check"
	"github.com/jack-sneddon/backup-butler/internal/commands/sync" // Updated import
	"github.com/jack-sneddon/backup-butler/internal/commands/version"
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

// Add to init()
func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")

	// Add commands
	rootCmd.AddCommand(version.NewVersionCmd())
	rootCmd.AddCommand(check.NewCheckCmd())
	rootCmd.AddCommand(sync.NewSyncCmd())
}
