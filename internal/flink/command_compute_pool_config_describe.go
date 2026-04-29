package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newComputePoolConfigDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe",
		Short:       "Describe a Flink compute pool config.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		RunE:        c.computePoolConfigDescribe,
	}

	pcmd.AddOutputFlag(cmd)
	return cmd
}

func (c *command) computePoolConfigDescribe(cmd *cobra.Command, args []string) error {
	computePoolConfig, err := c.V2Client.DescribeFlinkComputePoolConfig()
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&computePoolConfigOut{
		DefaultPoolEnabled: computePoolConfig.Spec.GetDefaultPoolEnabled(),
		DefaultPoolMaxCFU:  computePoolConfig.Spec.GetDefaultPoolMaxCfu(),
	})
	return table.Print()
}
