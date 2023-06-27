package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newIamBindingCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Flink IAM binding.",
		Args:  cobra.NoArgs,
		RunE:  c.iamBindingCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a Flink IAM binding for AWS region "us-west-2" and environment "env-123".`,
				Code: "confluent flink iam-binding create --cloud aws --region us-west-2 --environment env-123 --identity-pool pool-123",
			},
		),
	}

	c.addRegionFlag(cmd)
	pcmd.AddCloudFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("identity-pool", "", "Identity pool ID.")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) iamBindingCreate(cmd *cobra.Command, _ []string) error {
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
		if c.Context.GetCurrentIdentityPool() == "" {
			return errors.NewErrorWithSuggestions("no identity pool set", "Set a persistent identity pool with `confluent iam pool use` or pass the `--identity-pool` flag.")
		}
		identityPoolId = c.Context.GetCurrentIdentityPool()
	}

	iamBinding, err := c.V2Client.CreateFlinkIAMBinding(region, cloud, environmentId, identityPoolId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&iamBindingOut{
		Id:           iamBinding.GetId(),
		Region:       iamBinding.GetRegion(),
		Cloud:        iamBinding.GetCloud(),
		Environment:  iamBinding.Environment.GetId(),
		IdentityPool: iamBinding.GetIdentityPool().Id,
	})
	return table.Print()
}
