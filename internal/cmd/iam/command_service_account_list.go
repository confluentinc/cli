package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *serviceAccountCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service accounts.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *serviceAccountCommand) list(cmd *cobra.Command, _ []string) error {
	serviceAccounts, err := c.V2Client.ListIamServiceAccounts()
	if err != nil {
		return err
	}

	list := output.NewList(cmd, resource.ServiceAccount)

	for _, serviceAccount := range serviceAccounts {
		list.Add(&serviceAccountOut{
			ResourceId:  serviceAccount.GetId(),
			Name:        serviceAccount.GetDisplayName(),
			Description: serviceAccount.GetDescription(),
		})
	}

	return list.Print()
}
