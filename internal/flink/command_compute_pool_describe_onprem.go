package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newComputePoolDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe <name>",
		Short:       "Describe a Flink compute pool in Confluent Platform.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.computePoolDescribeOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) computePoolDescribeOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	sdkComputePool, err := client.DescribeComputePool(c.createContext(), environment, name)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		var creationTime string
		if sdkComputePool.GetMetadata().CreationTimestamp != nil {
			creationTime = *sdkComputePool.GetMetadata().CreationTimestamp
		}
		table.Add(&computePoolOutOnPrem{
			CreationTime: creationTime,
			Name:         sdkComputePool.GetMetadata().Name,
			Type:         sdkComputePool.GetSpec().Type,
			Phase:        sdkComputePool.GetStatus().Phase,
		})
		return table.Print()
	}

	localPool := convertSdkComputePoolToLocalComputePool(sdkComputePool)
	return output.SerializedOutput(cmd, localPool)
}
