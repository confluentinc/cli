package organization

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Confluent Cloud organizations.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	organizations, err := c.V2Client.ListOrgOrganizations()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, organization := range organizations {
		list.Add(&out{
			IsCurrent: organization.GetId() == c.Context.GetCurrentOrganization(),
			Id:        organization.GetId(),
			Name:      organization.GetDisplayName(),
		})
	}
	return list.Print()
}
