package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	prerunner pcmd.PreRunner
}

// New returns the default command object for interacting with RBAC.
func New(prerunner pcmd.PreRunner, config *config.Config) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedWithMDSCLICommand(
		&cobra.Command{
			Use:   "iam",
			Short: "Manage RBAC and IAM permissions.",
			Long:  "Manage Role Based Access (RBAC) and Identity and Access Management (IAM) permissions.",
		},
		config, prerunner)
	cmd := &command{
		AuthenticatedCLICommand: cliCmd,
		prerunner:  prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewRoleCommand(c.Config.Config, c.prerunner))
	c.AddCommand(NewRolebindingCommand(c.Config.Config, c.prerunner))
}
