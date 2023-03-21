package organization

import (
	"github.com/spf13/cobra"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the current Confluent Cloud organization.",
		Args:  cobra.NoArgs,
		RunE:  c.update,
	}

	cmd.Flags().String("name", "", "Name of the Confluent Cloud organization.")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	id := c.Context.GetOrganization().GetResourceId()

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	updateOrganization := orgv2.OrgV2Organization{DisplayName: orgv2.PtrString(name)}
	organization, httpResp, err := c.V2Client.UpdateOrgOrganization(id, updateOrganization)
	if err != nil {
		return errors.CatchOrgV2ResourceNotFoundError(err, resource.Organization, httpResp)
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		IsCurrent: organization.GetId() == id,
		Id:        organization.GetId(),
		Name:      organization.GetDisplayName(),
	})
	return table.Print()
}
