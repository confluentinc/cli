package flink

import (
	"github.com/spf13/cobra"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *computePoolCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "create <name>",
		Short:  "Create a Flink compute pool.",
		Args:   cobra.ExactArgs(1),
		RunE:   c.create,
		Hidden: true, // TODO: Remove for GA
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *computePoolCommand) create(cmd *cobra.Command, args []string) error {
	computePool := flinkv2.FcpmV2ComputePool{
		Spec: &flinkv2.FcpmV2ComputePoolSpec{
			DisplayName: flinkv2.PtrString(args[0]),
		},
	}

	computePool, err := c.V2Client.CreateFlinkComputePool(computePool)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		Id: computePool.GetId(),
	})
	return table.Print()
}
