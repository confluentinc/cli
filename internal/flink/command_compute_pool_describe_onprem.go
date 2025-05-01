package flink

import (
	"context"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newComputePoolDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe <name>",
		Short:       "Describe a Flink Compute Pool.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.computePoolDescribeOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the environment to describe the Flink Compute Pool from.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlagWithHumanRestricted(cmd)
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) computePoolDescribeOnPrem(cmd *cobra.Command, args []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	name := args[0]
	// Get the context from the command
	ctx := context.WithValue(context.Background(), cmfsdk.ContextAccessToken, c.Context.GetAuthToken())

	computePool, err := client.DescribeComputePool(ctx, environment, name)
	if err != nil {
		return err
	}

	return output.SerializedOutput(cmd, computePool)
}
