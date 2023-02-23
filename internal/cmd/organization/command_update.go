package organization

import (
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the current Confluent Cloud organization.",
		Args:  cobra.NoArgs,
		RunE:  c.update,
	}

	cmd.Flags().String("name", "", "Name of the Confluent Cloud organization.")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	id := c.Context.GetOrganization().GetResourceId()

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	updateOrganization := orgv2.OrgV2Organization{DisplayName: orgv2.PtrString(name)}
	_, httpResp, err := c.V2Client.UpdateOrgOrganization(id, updateOrganization)
	if err != nil {
		return errors.CatchOrgV2ResourceNotFoundError(err, resource.Organization, httpResp)
	}

	utils.ErrPrintf(cmd, errors.UpdateSuccessMsg, "name", "organization", id, name)
	return nil
}
