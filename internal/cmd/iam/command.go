package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	prerunner pcmd.PreRunner
	config    *v3.Config
}

// New returns the default command object for interacting with RBAC.
func New(prerunner pcmd.PreRunner, config *v3.Config) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedWithMDSCLICommand(
		&cobra.Command{
			Use:   "iam",
			Short: "Manage RBAC, ACL and IAM permissions.",
			Long:  "Manage Role Based Access (RBAC), Access Control Lists (ACL), and Identity and Access Management (IAM) permissions.",
		},
		config, prerunner)
	cmd := &command{
		AuthenticatedCLICommand: cliCmd,
		prerunner:               prerunner,
		config:                  config,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewRoleCommand(c.config, c.prerunner))
	c.AddCommand(NewRolebindingCommand(c.config, c.prerunner))
	c.AddCommand(NewACLCommand(c.config, c.prerunner))
}
