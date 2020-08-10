package admin

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func New(prerunner pcmd.PreRunner) *cobra.Command {
	c := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "admin",
			Short: "Perform admin-specific tasks.",
			Args:  cobra.NoArgs,
		},
		prerunner,
	)

	c.Hidden = true
	// TODO: payment command

	return c.Command
}
