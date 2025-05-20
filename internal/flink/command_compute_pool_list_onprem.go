package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newComputePoolListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List Flink Compute Pools in Confluent Platform.",
		Args:        cobra.NoArgs,
		RunE:        c.computePoolListOnPrem,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) computePoolListOnPrem(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	computePools, err := client.ListComputePools(c.createContext(), environment)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, pool := range computePools {
			list.Add(&computePoolOutOnPrem{
				CreationTime: pool.Metadata.GetCreationTimestamp(),
				Name:         pool.Metadata.Name,
				Type:         pool.Spec.Type,
				Phase:        pool.Status.Phase,
			})
		}
		return list.Print()
	}

	return output.SerializedOutput(cmd, computePools)
}
