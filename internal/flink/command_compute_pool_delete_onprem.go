package flink

import (
	"context"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newComputePoolDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "delete <name-1> [name-2] ... [name-n]",
		Short:       "Delete one or more Flink Compute Pools.",
		Args:        cobra.MinimumNArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.computePoolDeleteOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the environment to delete the Flink Compute Pool from.")
	addCmfFlagSet(cmd)
	pcmd.AddForceFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) computePoolDeleteOnPrem(cmd *cobra.Command, args []string) error {
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

	existenceFunc := func(name string) bool {
		_, err := client.DescribeComputePool(ctx, environment, name)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.FlinkComputePool); err != nil {
		// We are validating only the existence of the resources (there is no prefix validation).
		// Thus, we can add some extra context for the error.
		suggestions := "List available Flink compute pools with `confluent flink compute-pool list`."
		suggestions += "\nCheck that CMF is running and accessible."
		return errors.NewErrorWithSuggestions(err.Error(), suggestions)
	}

	deleteFunc := func(name string) error {
		return client.DeleteComputePool(ctx, environment, name)
	}

	_, err = deletion.Delete(args, deleteFunc, resource.FlinkComputePool)
	return err
}
