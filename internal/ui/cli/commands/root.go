// internal/ui/cli/commands/root.go
package commands

import (
	"flag"
	"fmt"

	"github.com/jack-sneddon/backup-butler/internal/app"
	"github.com/jack-sneddon/backup-butler/internal/ui/cli/formatter"
)

type CLI struct {
	configPath    string
	helpFlag      bool
	validateFlag  bool
	dryRunFlag    bool
	listVersions  bool
	showVersion   string
	latestVersion bool
	formatter     *formatter.OutputFormatter
}

func NewCLI() *CLI {
	return &CLI{
		formatter: formatter.NewOutputFormatter(),
	}
}

func (c *CLI) ParseFlags() {
	flag.StringVar(&c.configPath, "config", "", "Path to the configuration file")
	flag.BoolVar(&c.helpFlag, "help", false, "Show help message")
	flag.BoolVar(&c.validateFlag, "validate", false, "Validate the configuration file")
	flag.BoolVar(&c.dryRunFlag, "dry-run", false, "Simulate the backup process")
	flag.BoolVar(&c.listVersions, "list-versions", false, "List all backup versions")
	flag.StringVar(&c.showVersion, "show-version", "", "Show details of a specific backup version")
	flag.BoolVar(&c.latestVersion, "latest-version", false, "Show most recent backup details")

	flag.Parse()
}

func (c *CLI) Execute() int {
	if c.helpFlag {
		fmt.Println(c.formatter.FormatHelp())
		return 0
	}

	if c.configPath == "" {
		fmt.Println("Error: -config flag is required.")
		fmt.Println(c.formatter.FormatHelp())
		return 1
	}

	if c.validateFlag {
		fmt.Println("Configuration is valid.")
		return 0
	}

	// Create backup service
	factory := app.NewFactory(c.configPath)
	service, err := factory.CreateBackupService()
	if err != nil {
		fmt.Println(c.formatter.FormatError(err))
		return 1
	}

	// Handle version operations
	if c.listVersions {
		versions := service.GetVersions()
		fmt.Println(c.formatter.FormatVersionList(versions))
		return 0
	}

	if c.showVersion != "" {
		version, err := service.GetVersion(c.showVersion)
		if err != nil {
			fmt.Println(c.formatter.FormatError(err))
			return 1
		}
		fmt.Println(c.formatter.FormatVersionDetails(version))
		return 0
	}

	if c.latestVersion {
		version, err := service.GetLatestVersion()
		if err != nil {
			fmt.Println(c.formatter.FormatError(err))
			return 1
		}
		fmt.Println(c.formatter.FormatVersionDetails(version))
		return 0
	}

	// Handle backup operations
	backupCmd := NewBackupCommand(service, c.formatter)
	if c.dryRunFlag {
		return backupCmd.DryRun()
	}
	return backupCmd.Backup()
}
