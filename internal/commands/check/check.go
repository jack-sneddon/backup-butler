// internal/commands/check/check.go
package check

import "github.com/spf13/cobra"

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

func runCheck(cmd *cobra.Command, args []string) error {
	// TODO: Implement check logic
	return nil
}
