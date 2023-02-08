package flink

import (
	"github.com/spf13/cobra"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink/v2"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *resourcePoolCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Flink resource pool.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *resourcePoolCommand) create(cmd *cobra.Command, args []string) error {
	resourcePool := flinkv2.FrpmV2ResourcePool{
		Spec: &flinkv2.FrpmV2ResourcePoolSpec{
			DisplayName: flinkv2.PtrString(args[0]),
			// TODO: Cloud
			// TODO: Region
			// TODO: CSUs
		},
	}

	resourcePool, err := c.V2Client.CreateFlinkResourcePool(resourcePool)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		Id: resourcePool.GetId(),
	})
	return table.Print()
}
