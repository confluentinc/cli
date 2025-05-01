package flink

import (
	"context"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newComputePoolListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List Flink Compute Pools.",
		Args:        cobra.NoArgs,
		RunE:        c.computePoolListOnPrem,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.Flags().String("environment", "", "Name of the environment to list Flink Compute Pools from.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlagWithHumanRestricted(cmd)
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

	// Get the context from the command
	ctx := context.WithValue(context.Background(), cmfsdk.ContextAccessToken, c.Context.GetAuthToken())

	computePools, err := client.ListComputePools(ctx, environment)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, pool := range computePools {
			envInPool, ok := pool.Spec.ClusterSpec["environment"].(string)
			if !ok {
				envInPool = environment
			}
			list.Add(&computePoolOutOnPrem{
				Name:        pool.Metadata.Name,
				ID:          pool.Metadata.Uid,
				Environment: envInPool,
				Phase:       pool.Status.Phase,
			})
		}
		return list.Print()
	}

	return output.SerializedOutput(cmd, computePools)
}
