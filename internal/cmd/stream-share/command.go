package stream_share

import (
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/cobra"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	prerunner       pcmd.PreRunner
	analyticsClient analytics.Client
}

// New returns the default command object to perform operations on stream share.
func New(prerunner pcmd.PreRunner, analyticsClient analytics.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream-share",
		Short: "Manage stream share.",
		Long:  "Create and redeem shared token for a stream share.",
	}

	c := &command{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner),
		prerunner:               prerunner,
		analyticsClient:         analyticsClient,
	}
	c.init()

	return c.Command
}

func (c *command) init() {
	c.AddCommand(NewSharedTokenCommand(c.prerunner, c.analyticsClient).Command)
}
