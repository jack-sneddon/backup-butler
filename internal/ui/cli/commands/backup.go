// internal/ui/cli/commands/backup.go
package commands

import (
	"context"
	"fmt"

	"github.com/jack-sneddon/backup-butler/internal/domain/backup"
	"github.com/jack-sneddon/backup-butler/internal/ui/cli/formatter"
)

type BackupCommand struct {
	service   backup.BackupService
	formatter *formatter.OutputFormatter
}

func NewBackupCommand(service backup.BackupService, formatter *formatter.OutputFormatter) *BackupCommand {
	return &BackupCommand{
		service:   service,
		formatter: formatter,
	}
}

func (c *BackupCommand) Backup() int {
	ctx := context.Background()
	if err := c.service.Backup(ctx); err != nil {
		fmt.Println(c.formatter.FormatError(err))
		return 1
	}
	return 0
}

func (c *BackupCommand) DryRun() int {
	ctx := context.Background()
	if err := c.service.DryRun(ctx); err != nil {
		fmt.Println(c.formatter.FormatError(err))
		return 1
	}
	return 0
}
