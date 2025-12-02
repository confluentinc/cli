package organization

import (
	"github.com/spf13/cobra"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the current Confluent Cloud organization.",
		Args:  cobra.NoArgs,
		RunE:  c.update,
	}

	cmd.Flags().String("name", "", "Name of the Confluent Cloud organization.")
	cmd.Flags().Bool("jit-enabled", false, "Toggle Just-In-Time (JIT) user provisioning for SSO-enabled organizations.")
	cmd.Flags().Bool("scim-enabled", false, "Toggle SCIM user provisioning for SSO-enabled organizations.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) update(cmd *cobra.Command, _ []string) error {
	organization := orgv2.OrgV2Organization{}
	if cmd.Flags().Changed("jit-enabled") {
		jitEnabled, err := cmd.Flags().GetBool("jit-enabled")
		if err != nil {
			return err
		}
		organization.JitEnabled = orgv2.PtrBool(jitEnabled)
	}

	if cmd.Flags().Changed("scim-enabled") {
		scimEnabled, err := cmd.Flags().GetBool("scim-enabled")
		if err != nil {
			return err
		}
		organization.ScimEnabled = orgv2.PtrBool(scimEnabled)
	}

	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		organization.DisplayName = orgv2.PtrString(name)
	}

	organization, httpResp, err := c.V2Client.UpdateOrgOrganization(c.Context.GetCurrentOrganization(), organization)
	if err != nil {
		return errors.CatchCCloudV2ResourceNotFoundError(err, resource.Organization, httpResp)
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		IsCurrent:   organization.GetId() == c.Context.GetCurrentOrganization(),
		Id:          organization.GetId(),
		Name:        organization.GetDisplayName(),
		JitEnabled:  organization.GetJitEnabled(),
		ScimEnabled: organization.GetScimEnabled(),
	})
	return table.Print()
}
