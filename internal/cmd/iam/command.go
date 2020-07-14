package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	prerunner pcmd.PreRunner
}

// New returns the default command object for interacting with RBAC.
func New(cliName string, prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedWithMDSCLICommand(
		&cobra.Command{
			Use:   "iam",
			Short: "Manage RBAC, ACL and IAM permissions.",
			Long:  "Manage Role-Based Access Control (RBAC), Access Control Lists (ACL), and Identity and Access Management (IAM) permissions.",
		}, prerunner)

	c := &command{
		AuthenticatedCLICommand: cliCmd,
		prerunner:               prerunner,
	}

	c.AddCommand(NewRoleCommand(c.prerunner))
	c.AddCommand(NewRolebindingCommand(c.prerunner))
	c.AddCommand(NewACLCommand(cliName, c.prerunner))

	return c.Command
}
