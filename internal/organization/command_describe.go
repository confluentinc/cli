package organization

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

type out struct {
	IsCurrent  bool   `human:"Current" serialized:"is_current"`
	Id         string `human:"ID" serialized:"id"`
	Name       string `human:"Name" serialized:"name"`
	JitEnabled bool   `human:"JIT Enabled" serialized:"jit_enabled"`
}

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe the current Confluent Cloud organization.",
		Args:  cobra.NoArgs,
		RunE:  c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, _ []string) error {
	organization, httpResp, err := c.V2Client.GetOrgOrganization(c.Context.GetCurrentOrganization())
	if err != nil {
		return errors.CatchCCloudV2ResourceNotFoundError(err, resource.Organization, httpResp)
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		IsCurrent:  organization.GetId() == c.Context.GetCurrentOrganization(),
		Id:         organization.GetId(),
		Name:       organization.GetDisplayName(),
		JitEnabled: organization.GetJitEnabled(),
	})
	return table.Print()
}
