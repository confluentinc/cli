package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/featureflags"
)

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "iam",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	if cfg.IsOnPremLogin() {
		cmd.Short = "Manage RBAC, ACL and IAM permissions."
		cmd.Long = "Manage Role-Based Access Control (RBAC), Access Control Lists (ACL), and Identity and Access Management (IAM) permissions."
	} else {
		cmd.Short = "Manage RBAC and IAM permissions."
		cmd.Long = "Manage Role-Based Access Control (RBAC) and Identity and Access Management (IAM) permissions."
	}

	_ = cfg.ParseFlagsIntoConfig(cmd)
	if cfg.IsTest || featureflags.Manager.BoolVariation("cli.iam.group_mapping.enable", cfg.Context(), config.CliLaunchDarklyClient, true, false) {
		cmd.AddCommand(newGroupMappingCommand(prerunner))
	}
	cmd.AddCommand(newAclCommand(prerunner))
	cmd.AddCommand(newIpFilterCommand(prerunner))
	cmd.AddCommand(newIpGroupCommand(prerunner))
	cmd.AddCommand(newPoolCommand(prerunner))
	cmd.AddCommand(newProviderCommand(prerunner))
	cmd.AddCommand(newRbacCommand(cfg, prerunner))
	cmd.AddCommand(newServiceAccountCommand(prerunner))
	cmd.AddCommand(newUserCommand(cfg, prerunner))

	return cmd
}
