package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *serviceAccountCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service accounts.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().StringSlice("display-name", nil, "A comma-separated list of service account display names.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *serviceAccountCommand) list(cmd *cobra.Command, _ []string) error {
	name, err := cmd.Flags().GetStringSlice("display-name")
	if err != nil {
		return err
	}

	serviceAccounts, err := c.V2Client.ListIamServiceAccounts(name)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, serviceAccount := range serviceAccounts {
		list.Add(&serviceAccountOut{
			ResourceId:  serviceAccount.GetId(),
			Name:        serviceAccount.GetDisplayName(),
			Description: serviceAccount.GetDescription(),
		})
	}
	return list.Print()
}
