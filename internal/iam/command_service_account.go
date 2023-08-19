package iam

import (
	"fmt"

	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type serviceAccountCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type serviceAccountOut struct {
	ResourceId  string `human:"ID" serialized:"id"`
	Name        string `human:"Name" serialized:"name"`
	Description string `human:"Description" serialized:"description"`
}

func newServiceAccountCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "service-account",
		Aliases:     []string{"sa"},
		Short:       "Manage service accounts.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &serviceAccountCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printServiceAccount(cmd *cobra.Command, serviceAccount iamv2.IamV2ServiceAccount) error {
	table := output.NewTable(cmd)
	table.Add(&serviceAccountOut{
		ResourceId:  serviceAccount.GetId(),
		Name:        serviceAccount.GetDisplayName(),
		Description: serviceAccount.GetDescription(),
	})
	return table.Print()
}

func (c *serviceAccountCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *serviceAccountCommand) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteServiceAccounts(c.V2Client)
}

func requireLen(val string, maxLen int, field string) error {
	if len(val) > maxLen {
		return fmt.Errorf("%s length should not exceed %d characters", field, maxLen)
	}

	return nil
}
