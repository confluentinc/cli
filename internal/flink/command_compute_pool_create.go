package flink

import (
	"github.com/spf13/cobra"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2/flink/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newComputePoolCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Flink compute pool.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.computePoolCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create Flink compute pool "my-compute-pool" in AWS with 5 CFUs.`,
				Code: "confluent flink compute-pool create my-compute-pool --cloud aws --region us-west-2 --max-cfu 5",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().Int32("max-cfu", 5, "Maximum number of Confluent Flink Units (CFU).")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) computePoolCreate(cmd *cobra.Command, args []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	maxCfu, err := cmd.Flags().GetInt32("max-cfu")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	environment, err := c.V2Client.GetOrgEnvironment(environmentId)
	if err != nil {
		return err
	}

	computePool := flinkv2.FcpmV2ComputePool{Spec: &flinkv2.FcpmV2ComputePoolSpec{
		DisplayName: flinkv2.PtrString(args[0]),
		Cloud:       flinkv2.PtrString(cloud),
		Region:      flinkv2.PtrString(region),
		MaxCfu:      flinkv2.PtrInt32(maxCfu),
		Environment: &flinkv2.GlobalObjectReference{
			Id:           environmentId,
			Related:      environment.Metadata.GetSelf(),
			ResourceName: environment.Metadata.GetResourceName(),
		},
	}}

	computePool, err = c.V2Client.CreateFlinkComputePool(computePool)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&computePoolOut{
		IsCurrent:  computePool.GetId() == c.Context.GetCurrentFlinkComputePool(),
		Id:         computePool.GetId(),
		Name:       computePool.Spec.GetDisplayName(),
		CurrentCfu: computePool.Status.GetCurrentCfu(),
		MaxCfu:     computePool.Spec.GetMaxCfu(),
		Region:     computePool.Spec.GetRegion(),
		Status:     computePool.Status.GetPhase(),
	})
	return table.Print()
}
