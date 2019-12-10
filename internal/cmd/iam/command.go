package iam

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/version"
)

type command struct {
	*pcmd.CLICommand
	prerunner pcmd.PreRunner
}

// New returns the default command object for interacting with RBAC.
func New(prerunner pcmd.PreRunner, config *config.Config,
	version *version.Version, client *mds.APIClient) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "iam",
			Short: "Manage RBAC and IAM permissions.",
			Long:  "Manage Role Based Access (RBAC) and Identity and Access Management (IAM) permissions.",
		},
		config, prerunner)
	cmd := &command{
		CLICommand: cliCmd,
		prerunner:  prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *command) init() {
	c.AddCommand(NewRoleCommand(c.Config, c.prerunner))
	c.AddCommand(NewRolebindingCommand(c.Config, c.prerunner))
}
