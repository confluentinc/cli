package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newArtifactDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "delete <name-1> [name-2] ... [name-n]",
		Short:       "Delete one or more Flink artifacts in Confluent Platform.",
		Long:        "Delete one or more Flink artifacts in Confluent Platform, including all of their versions.",
		Args:        cobra.MinimumNArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.artifactDeleteOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddForceFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) artifactDeleteOnPrem(cmd *cobra.Command, args []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	existenceFunc := func(name string) bool {
		_, err := client.DescribeArtifact(c.createContext(), environment, name, "")
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.FlinkArtifact); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), artifactLookupSuggestions)
	}

	// An empty version deletes the artifact and all of its versions.
	deleteFunc := func(name string) error {
		return client.DeleteArtifact(c.createContext(), environment, name, "")
	}

	_, err = deletion.Delete(cmd, args, deleteFunc, resource.FlinkArtifact)
	return err
}
