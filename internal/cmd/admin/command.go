package admin

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func New(prerunner pcmd.PreRunner, isTest bool) *cobra.Command {
	c := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "admin",
			Short: "Perform administrative tasks for the current organization.",
			Args:  cobra.NoArgs,
		},
		prerunner,
	)

	c.AddCommand(NewPaymentCommand(prerunner, isTest))

	return c.Command
}
