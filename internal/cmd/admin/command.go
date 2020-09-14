package admin

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func New(prerunner pcmd.PreRunner, isTest bool) *cobra.Command {
	c := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "admin", // TODO: rename to org?
			Short: "Perform admin-specific tasks.",
			Args:  cobra.NoArgs,
		},
		prerunner,
	)

	c.AddCommand(NewPaymentCommand(prerunner, isTest))
	c.AddCommand(NewUsersCommand(prerunner))

	return c.Command
}
