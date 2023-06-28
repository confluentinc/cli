package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newIamBindingListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink IAM bindings.",
		Args:  cobra.NoArgs,
		RunE:  c.iamBindingList,
	}

	pcmd.AddCloudFlag(cmd)
	c.addRegionFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("identity-pool", "", "Identity pool ID.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) iamBindingList(cmd *cobra.Command, _ []string) error {
	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	identityPoolId, err := cmd.Flags().GetString("identity-pool")
	if err != nil {
		return err
	}
	if identityPoolId == "" {
		identityPoolId = c.Context.GetCurrentIdentityPool()
	}

	iamBindings, err := c.V2Client.ListFlinkIAMBindings(environmentId, region, cloud, identityPoolId)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, iamBinding := range iamBindings {
		list.Add(&iamBindingOut{
			Id:           iamBinding.GetId(),
			Cloud:        iamBinding.GetCloud(),
			Region:       iamBinding.GetRegion(),
			Environment:  iamBinding.GetEnvironment().Id,
			IdentityPool: iamBinding.GetIdentityPool().Id,
		})
	}
	return list.Print()
}
