package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/utils"
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

	cmd.AddCommand(newAclCommand(prerunner))
	cmd.AddCommand(newCertificateAuthorityCommand(cfg, prerunner))
	cmd.AddCommand(newCertificatePoolCommand(cfg, prerunner))
	cmd.AddCommand(newGroupMappingCommand(prerunner))
	cmd.AddCommand(newIpFilterCommand(cfg, prerunner))
	cmd.AddCommand(newIpGroupCommand(prerunner))
	cmd.AddCommand(newPoolCommand(cfg, prerunner))
	cmd.AddCommand(newProviderCommand(prerunner))
	cmd.AddCommand(newRbacCommand(cfg, prerunner))
	cmd.AddCommand(newServiceAccountCommand(cfg, prerunner))
	cmd.AddCommand(newUserCommand(cfg, prerunner))

	return cmd
}

func addResourceOwnerFlag(cmd *cobra.Command, cliCommand *pcmd.AuthenticatedCLICommand) {
	items := []string{"user", "group-mapping", "service-account", "identity-pool"}
	description := fmt.Sprintf("The resource ID of the principal who will be assigned resource owner on the "+
		"created resource. Principal can be a %s.", utils.ArrayToCommaDelimitedString(items, "or"))
	cmd.Flags().String("resource-owner", "", description)
	pcmd.RegisterFlagCompletionFunc(cmd, "resource-owner", func(cmd *cobra.Command, args []string) []string {
		if err := cmd.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return autocompleteResourceOwners(cliCommand.V2Client)
	})
}
func autocompleteResourceOwners(client *ccloudv2.Client) []string {
	users, err := client.ListIamUsers()
	if err != nil {
		return nil
	}
	groupMappings, err := client.ListGroupMappings()
	if err != nil {
		return nil
	}
	serviceAccounts, err := client.ListIamServiceAccounts()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(users)+len(groupMappings)+len(serviceAccounts))
	offset := 0
	for i, user := range users {
		description := user.GetFullName()
		suggestions[i] = fmt.Sprintf("%s\t%s", user.GetId(), description)
	}
	offset += len(users)
	for i, groupMapping := range groupMappings {
		description := fmt.Sprintf("%s: %s", groupMapping.GetDisplayName(), groupMapping.GetDescription())
		suggestions[i+offset] = fmt.Sprintf("%s\t%s", groupMapping.GetId(), description)
	}
	offset += len(groupMappings)
	for i, serviceAccount := range serviceAccounts {
		description := fmt.Sprintf("%s: %s", serviceAccount.GetDisplayName(), serviceAccount.GetDisplayName())
		suggestions[i+offset] = fmt.Sprintf("%s\t%s", serviceAccount.GetId(), description)
	}
	return suggestions
}
