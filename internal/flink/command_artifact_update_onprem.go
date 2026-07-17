package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newArtifactUpdateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "update <name>",
		Short:       "Update a Flink artifact's metadata in Confluent Platform.",
		Long:        "Update the labels of a Flink artifact in Confluent Platform without uploading new content.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.artifactUpdateOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Replace the labels of Flink artifact "my-artifact" in the environment "my-environment".`,
				Code: "confluent flink artifact update my-artifact --label owner=team-a,tier=gold --environment my-environment",
			},
		),
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().StringSlice("label", nil, `A comma-separated list of "key=value" label pairs. Provide the complete set of labels to apply; omit the flag to leave existing labels unchanged.`)
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) artifactUpdateOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	// Only set labels when --label is provided. Leaving them nil omits the field so existing labels are preserved server-side.
	var labels map[string]string
	if cmd.Flags().Changed("label") {
		labels, err = getLabelsFlag(cmd)
		if err != nil {
			return err
		}
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	// Metadata-only update: no file is uploaded, so no new version is created.
	outputArtifact, err := client.UpdateArtifact(c.createContext(), environment, name, newSdkArtifact(name, labels), nil)
	if err != nil {
		return err
	}

	return printArtifactOnPrem(cmd, outputArtifact)
}
