package admin

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/log"
)

func New(prerunner pcmd.PreRunner, logger *log.Logger, userAgent string) *cobra.Command {
	c := pcmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "admin",
			Short: "Perform admin-specific tasks.",
			Args:  cobra.NoArgs,
		},
		prerunner,
	)

	c.AddCommand(NewSignupCommand(prerunner, logger, userAgent))
	// TODO: payment command

	return c.Command
}
