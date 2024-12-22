// cmd/main.go
package main

import (
	"os"

	"github.com/jack-sneddon/backup-butler/internal/ui/cli/commands"
)

func main() {
	cli := commands.NewCLI()
	cli.ParseFlags()
	os.Exit(cli.Execute())
}
