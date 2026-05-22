package flink

import (
	"github.com/spf13/cobra"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2/flink/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newComputePoolConfigUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update",
		Short:             "Update a Flink compute pool config.",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validComputePoolArgs),
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		RunE:              c.computePoolConfigUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update CFU count of a Flink compute pool config.`,
				Code: `confluent flink compute-pool update --max-cfu 5`,
			},
		),
	}

	cmd.Flags().Int32("max-cfu", -1, "Maximum number of Confluent Flink Units (CFUs) that default compute pools in this organization should auto-scale to.")
	cmd.Flags().Bool("default-pool", false, "Whether default compute pools are enabled for the organization.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsOneRequired("max-cfu", "default-pool")

	return cmd
}

func (c *command) computePoolConfigUpdate(cmd *cobra.Command, args []string) error {
	computePoolConfig, err := c.V2Client.DescribeFlinkComputePoolConfig()
	if err != nil {
		return err
	}

	update := flinkv2.FcpmV2OrgComputePoolConfigUpdate{
		Spec: &flinkv2.FcpmV2OrgComputePoolConfigSpec{
			DefaultPoolEnabled: flinkv2.PtrBool(computePoolConfig.Spec.GetDefaultPoolEnabled()),
			DefaultPoolMaxCfu:  flinkv2.PtrInt32(computePoolConfig.Spec.GetDefaultPoolMaxCfu()),
		},
	}

	if cmd.Flags().Changed("default-pool") {
		defaultPool, err := cmd.Flags().GetBool("default-pool")
		if err != nil {
			return err
		}
		update.Spec.DefaultPoolEnabled = flinkv2.PtrBool(defaultPool)
	}

	maxCfu, err := cmd.Flags().GetInt32("max-cfu")
	if err != nil {
		return err
	}
	if maxCfu != -1 {
		update.Spec.DefaultPoolMaxCfu = flinkv2.PtrInt32(maxCfu)
	}

	updatedComputePoolConfig, err := c.V2Client.UpdateFlinkComputePoolConfig(update)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&computePoolConfigOut{
		DefaultPoolEnabled: updatedComputePoolConfig.Spec.GetDefaultPoolEnabled(),
		DefaultPoolMaxCFU:  updatedComputePoolConfig.Spec.GetDefaultPoolMaxCfu(),
	})
	return table.Print()
}
