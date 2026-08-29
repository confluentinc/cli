package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
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

	cmd.AddCommand(
		newAclCommand(prerunner),
		newCertificateAuthorityCommand(prerunner),
		newCertificatePoolCommand(cfg, prerunner),
		newGroupMappingCommand(prerunner),
		newIpFilterCommand(cfg, prerunner),
		newIpGroupCommand(prerunner),
		newPoolCommand(cfg, prerunner),
		newProviderCommand(prerunner),
		newRbacCommand(cfg, prerunner),
		newServiceAccountCommand(cfg, prerunner),
		newUserCommand(cfg, prerunner),
		// cli-tfgen:cli-subcommands
	)

	return cmd
}
