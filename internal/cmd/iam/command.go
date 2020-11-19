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
	var cliCmd *pcmd.AuthenticatedCLICommand
	if cliName == "confluent" {
		cliCmd = pcmd.NewAuthenticatedWithMDSCLICommand(
			&cobra.Command{
				Use:   "iam",
				Short: "Manage RBAC, ACL and IAM permissions.",
				Long:  "Manage Role-Based Access Control (RBAC), Access Control Lists (ACL), and Identity and Access Management (IAM) permissions.",
			}, prerunner)
	} else {
		cliCmd = pcmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:   "iam",
				Short: "Manage RBAC and IAM permissions.",
				Long:  "Manage Role-Based Access Control (RBAC) and Identity and Access Management (IAM) permissions.",
			}, prerunner)
	}

	c := &command{
		AuthenticatedCLICommand: cliCmd,
		prerunner:               prerunner,
	}

	c.AddCommand(NewRoleCommand(cliName, c.prerunner))
	c.AddCommand(NewRolebindingCommand(cliName, c.prerunner))
	if cliName != "ccloud" {
		c.AddCommand(NewACLCommand(cliName, c.prerunner))
	}

	return c.Command
}
