package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	prerunner pcmd.PreRunner
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "iam",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	var cliCmd *pcmd.AuthenticatedCLICommand
	if cfg.IsOnPremLogin() {
		cmd.Short = "Manage RBAC, ACL and IAM permissions."
		cmd.Long = "Manage Role-Based Access Control (RBAC), Access Control Lists (ACL), and Identity and Access Management (IAM) permissions."
		cliCmd = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
	} else {
		cmd.Short = "Manage RBAC and IAM permissions."
		cmd.Long = "Manage Role-Based Access Control (RBAC) and Identity and Access Management (IAM) permissions."
		cliCmd = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)
	}

	c := &command{
		AuthenticatedCLICommand: cliCmd,
		prerunner:               prerunner,
	}

	dc := dynamicconfig.New(cfg, nil, nil)
	_ = dc.ParseFlagsIntoConfig(cmd)
	if cfg.IsTest || featureflags.Manager.BoolVariation("cli.identity-provider", dc.Context(), v1.CliLaunchDarklyClient, true, false) {
		c.AddCommand(newPoolCommand(c.prerunner))
		c.AddCommand(newProviderCommand(c.prerunner))
	}
	c.AddCommand(newACLCommand(c.prerunner))
	c.AddCommand(newRBACCommand(cfg, c.prerunner))
	c.AddCommand(newServiceAccountCommand(c.prerunner))
	c.AddCommand(newUserCommand(c.prerunner))

	return c.Command
}
