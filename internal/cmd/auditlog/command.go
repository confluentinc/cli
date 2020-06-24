package auditlog

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	prerunner pcmd.PreRunner
}

// New returns the default command object for interacting with audit logs.
func New(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedWithMDSCLICommand(
		&cobra.Command{
			Use:   "audit-log",
			Short: "Manage audit log configuration (since 6.0).",
			Long:  "Manage which auditable events are logged, and where the events are sent. (since 6.0).",
		}, prerunner)
	cmd := &command{
		AuthenticatedCLICommand: cliCmd,
		prerunner:               prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewConfigCommand(c.prerunner))
	c.AddCommand(NewRouteCommand(c.prerunner))
}
