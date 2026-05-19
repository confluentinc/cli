package debug

import (
	"github.com/spf13/cobra"
)

func (c *command) newPanicCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "panic",
		Short: "Trigger a test panic for crash reporting validation.",
		Long:  "Trigger a test panic to validate the crash reporting pipeline; note that panic traces are only collected and reported when logged in to Confluent Cloud.",
		Args:  cobra.NoArgs,
		RunE:  c.panic,
	}
}

func (c *command) panic(_ *cobra.Command, _ []string) error {
	panic("test panic")
}
