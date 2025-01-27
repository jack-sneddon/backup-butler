// internal/commands/root.go
package commands

import (
	"os"

	"github.com/jack-sneddon/backup-butler/internal/commands/check"
	"github.com/jack-sneddon/backup-butler/internal/commands/sync" // Updated import
	"github.com/jack-sneddon/backup-butler/internal/commands/version"
	"github.com/jack-sneddon/backup-butler/internal/config"
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
	// rootCmd.PersistentFlags().String("log-level", "", "Override config log level")
	rootCmd.SilenceUsage = true // Don't show usage on errors

	// 2. Initialize logger in PreRun
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// 1. Check command line flag
		if level := cmd.Flags().Lookup("log-level").Value.String(); level != "" {
			return logger.SetLevel(level)
		}

		// 2. Check config
		if cfgFile := cmd.Flags().Lookup("config").Value.String(); cfgFile != "" {
			cfg, err := config.LoadConfig(cfgFile)
			if err != nil {
				return err
			}
			if cfg.Logging.Level != "" {
				return logger.SetLevel(cfg.Logging.Level)
			}
		}

		// 3. Default to error
		return logger.SetLevel("error")
	}

	// 3. Add commands
	rootCmd.AddCommand(version.NewVersionCmd())
	rootCmd.AddCommand(check.NewCheckCmd())
	rootCmd.AddCommand(sync.NewSyncCmd())
}
