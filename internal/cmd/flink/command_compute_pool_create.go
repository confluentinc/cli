package flink

import (
	"github.com/spf13/cobra"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newComputePoolCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "create <name>",
		Short:  "Create a Flink compute pool.",
		Args:   cobra.ExactArgs(1),
		RunE:   c.create,
		Hidden: true, // TODO: Remove for GA
	}

	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("region", "", `Cloud region ID (use "confluent flink region list" to see all).`)
	pcmd.AddOutputFlag(cmd)

	pcmd.RegisterFlagCompletionFunc(cmd, "region", c.autocompleteRegions)

	return cmd
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	computePool := flinkv2.FcpmV2ComputePool{
		Spec: &flinkv2.FcpmV2ComputePoolSpec{
			DisplayName: flinkv2.PtrString(args[0]),
			Cloud:       flinkv2.PtrString(cloud),
			Region:      flinkv2.PtrString(region),
		},
	}

	computePool, err = c.V2Client.CreateFlinkComputePool(computePool)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&computePoolOut{
		Id:   computePool.GetId(),
		Name: computePool.Spec.GetDisplayName(),
	})
	return table.Print()
}
