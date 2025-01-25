// internal/commands/version/version.go
package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	GitCommit = "none"
	BuildTime = "unknown"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run:   runVersion,
	}
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("Backup Butler %s\n", Version)
	fmt.Printf("Git commit: %s\n", GitCommit)
	fmt.Printf("Built: %s\n", BuildTime)
}
