// internal/ui/cli/formatter/output.go
package formatter

import (
	"fmt"
	"strings"
	"time"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
)

const (
	dividerLine = "---------------"
	megabyte    = 1024 * 1024
)

type OutputFormatter struct{}

func NewOutputFormatter() *OutputFormatter {
	return &OutputFormatter{}
}

func (f *OutputFormatter) FormatHelp() string {
	return `backup-butler - Backup Utility

Usage:
  backup-butler [options]

Options:
  -config <file>       Path to the configuration file (JSON or YAML)
  --help, -h          Show this help message and exit
  --validate          Validate the configuration file without performing a backup
  --dry-run           Simulate the backup process without making any changes
  --list-versions     List all backup versions
  --show-version <id> Show details of a specific backup version
  --latest-version    Show most recent backup details

Examples:
  backup-butler -config backup_config.json
  backup-butler -config backup_config.yaml --dry-run
  backup-butler -config backup_config.yaml --list-versions
  backup-butler -config backup_config.yaml --show-version 20240117-150405
  backup-butler -config backup_config.yaml --latest-version`
}

func (f *OutputFormatter) FormatVersionList(versions []backup.BackupVersion) string {
	if len(versions) == 0 {
		return "No backup versions found"
	}

	var output strings.Builder
	output.WriteString("\nBackup History:\n")
	output.WriteString(dividerLine + "\n")

	for _, v := range versions {
		fmt.Fprintf(&output, "ID: %s\n", v.ID)
		fmt.Fprintf(&output, "  Time: %s\n", v.Timestamp.Format(time.RFC3339))
		fmt.Fprintf(&output, "  Duration: %v\n", v.Duration)
		fmt.Fprintf(&output, "  Files: %d total (%d copied, %d skipped, %d failed)\n",
			v.Stats.TotalFiles, v.Stats.FilesBackedUp, v.Stats.FilesSkipped, v.Stats.FilesFailed)
		fmt.Fprintf(&output, "  Size: %.2f MB\n", float64(v.Size)/megabyte)
		fmt.Fprintf(&output, "  Status: %s\n", v.Status)
		output.WriteString(dividerLine + "\n")
	}

	return output.String()
}

func (f *OutputFormatter) FormatVersionDetails(version *backup.BackupVersion) string {
	var output strings.Builder

	fmt.Fprintf(&output, "\nBackup Version Details: %s\n", version.ID)
	output.WriteString("-------------------------\n")
	fmt.Fprintf(&output, "Timestamp: %s\n", version.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(&output, "Duration: %v\n", version.Duration)
	fmt.Fprintf(&output, "Status: %s\n", version.Status)

	output.WriteString("\nStatistics:\n")
	fmt.Fprintf(&output, "  Total Files Processed: %d\n", version.Stats.TotalFiles)
	fmt.Fprintf(&output, "  Files Backed Up: %d\n", version.Stats.FilesBackedUp)
	fmt.Fprintf(&output, "  Files Skipped: %d\n", version.Stats.FilesSkipped)
	fmt.Fprintf(&output, "  Files Failed: %d\n", version.Stats.FilesFailed)
	fmt.Fprintf(&output, "  Total Size: %.2f MB\n", float64(version.Stats.TotalBytes)/megabyte)
	fmt.Fprintf(&output, "  Data Transferred: %.2f MB\n", float64(version.Stats.BytesTransferred)/megabyte)

	output.WriteString("\nConfiguration Used:\n")
	fmt.Fprintf(&output, "  Source Directory: %s\n", version.ConfigUsed.SourceDirectory)
	fmt.Fprintf(&output, "  Target Directory: %s\n", version.ConfigUsed.TargetDirectory)
	fmt.Fprintf(&output, "  Concurrency: %d\n", version.ConfigUsed.Concurrency)
	fmt.Fprintf(&output, "  Deep Duplicate Check: %v\n", version.ConfigUsed.DeepDuplicateCheck)

	return output.String()
}

func (f *OutputFormatter) FormatError(err error) string {
	return fmt.Sprintf("Error: %v", err)
}
