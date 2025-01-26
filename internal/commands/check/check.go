// internal/commands/check/check.go
package check

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jack-sneddon/backup-butler/internal/config"
	"github.com/jack-sneddon/backup-butler/internal/logger"
	"github.com/jack-sneddon/backup-butler/internal/scan"
	"github.com/spf13/cobra"
)

func NewCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [source] [target]",
		Short: "Check backup integrity",
		RunE:  runCheck,
	}

	cmd.Flags().StringP("level", "l", "standard", "validation level (quick|standard|deep)")
	cmd.Flags().StringP("output", "o", "text", "output format (text|csv|html)")

	return cmd
}

// internal/commands/check/check.go
func runCheck(cmd *cobra.Command, args []string) error {
	log := logger.Get()

	cfgFile := cmd.Root().PersistentFlags().Lookup("config").Value.String()
	log.Debugw("Loading config", "file", cfgFile)

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	scanner := scan.NewScanner()
	progress, err := scanner.Scan(cfg.Source)
	if err != nil {
		return err
	}

	log.Infow("Scan complete",
		"dirs", progress.ScannedDirs,
		"files", progress.ScannedFiles,
		"size", progress.TotalBytes)

	fmt.Printf("\nScan Results:\n")
	fmt.Printf("├── Source: %s\n", cfg.Source)
	fmt.Printf("├── Summary\n")
	fmt.Printf("│   ├── Directories: %d\n", progress.ScannedDirs)
	fmt.Printf("│   ├── Files: %d\n", progress.ScannedFiles)
	fmt.Printf("│   └── Total Size: %s\n", formatBytes(progress.TotalBytes))

	comparisons, err := scanner.Compare(cfg.Source, cfg.Target)
	if err != nil {
		return err
	}

	fmt.Printf("└── File Status\n")
	for _, comp := range comparisons {
		fmt.Printf("    %c %s\n", comp.Status, comp.Path)
	}

	return nil
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

func formatStats(scanner *scan.Scanner, rootPath string) string {
	log := logger.Get()
	log.Infow("Formatting stats", "root", rootPath)

	var b strings.Builder
	for _, dir := range scanner.GetDirectoryStats() {
		relPath, _ := filepath.Rel(rootPath, dir.Path)
		if relPath == "." {
			continue
		}
		b.WriteString(fmt.Sprintf("    ├── %s\n", relPath))
		b.WriteString(fmt.Sprintf("    │   ├── Files: %d\n", dir.FileCount))
		b.WriteString(fmt.Sprintf("    │   └── Size: %s\n", formatBytes(dir.TotalSize)))
	}
	return b.String()
}
