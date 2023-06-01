package flink

import (
	"strings"

	"github.com/spf13/cobra"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2/flink/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newComputePoolCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <id>",
		Short: "Create a Flink compute pool.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.computePoolCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create Flink compute pool "my-compute-pool" in AWS with 2 CFUs.`,
				Code: "confluent flink compute-pool create my-compute-pool --cloud aws --region us-west-2 --cfu 2",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("region", "", `Cloud region for compute pool (use "confluent flink region list" to see all).`)
	cmd.Flags().Int32("cfu", 1, "Number of Confluent Flink Units (CFU).")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	pcmd.RegisterFlagCompletionFunc(cmd, "region", c.autocompleteRegions)

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

	cfu, err := cmd.Flags().GetInt32("cfu")
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
		Config:      &flinkv2.FcpmV2ComputePoolSpecConfigOneOf{FcpmV2Standard: &flinkv2.FcpmV2Standard{Kind: "Standard"}},
		MaxCfu:      flinkv2.PtrInt32(cfu),
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
		IsCurrent: computePool.GetId() == c.Context.GetCurrentFlinkComputePool(),
		Id:        computePool.GetId(),
		Name:      computePool.Spec.GetDisplayName(),
		Cfu:       computePool.Spec.GetMaxCfu(),
		Region:    computePool.Spec.GetRegion(),
		Status:    computePool.Status.GetPhase(),
	})
	return table.Print()
}

func (c *command) autocompleteRegions(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return nil
	}

	regions, err := c.V2Client.ListFlinkRegions(strings.ToUpper(cloud))
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(regions))
	for i, region := range regions {
		suggestions[i] = region.GetRegionName()
	}
	return suggestions
}
