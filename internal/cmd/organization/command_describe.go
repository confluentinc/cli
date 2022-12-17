package organization

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

type out struct {
	IsCurrent bool   `human:"Current" serialized:"is_current"`
	Id        string `human:"ID" serialized:"id"`
	Name      string `human:"Name" serialized:"name"`
}

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a Confluent Cloud organization.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	organization, httpResp, err := c.V2Client.GetOrgOrganization(args[0])
	if err != nil {
		return errors.CatchOrgV2ResourceNotFoundError(err, resource.Organization, httpResp)
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		IsCurrent: *organization.Id == c.Context.GetOrganization().GetResourceId(),
		Id:        *organization.Id,
		Name:      *organization.DisplayName,
	})
	return table.Print()
}
