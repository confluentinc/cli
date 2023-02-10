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

	pcmd.AddContextFlag(cmd, c.CLICommand)
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
			IsCurrent: *organization.Id == c.Context.GetOrganization().GetResourceId(),
			Id:        *organization.Id,
			Name:      *organization.DisplayName,
		})
	}
	return list.Print()
}
