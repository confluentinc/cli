package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newArtifactCreateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "create <name>",
		Short:       "Create a Flink artifact in Confluent Platform.",
		Long:        "Create a Flink artifact in Confluent Platform by uploading a JAR or ZIP file. This creates version 1 of the artifact.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.artifactCreateOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create Flink artifact "my-artifact" in the environment "my-environment".`,
				Code: "confluent flink artifact create my-artifact --artifact-file artifact.jar --environment my-environment",
			},
		),
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("artifact-file", "", "Path to the Flink artifact JAR or ZIP file.")
	cmd.Flags().StringSlice("label", nil, `A comma-separated list of "key=value" label pairs.`)
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("artifact-file"))
	cobra.CheckErr(cmd.MarkFlagFilename("artifact-file", "jar", "zip"))

	return cmd
}

func (c *command) artifactCreateOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	artifactFile, err := cmd.Flags().GetString("artifact-file")
	if err != nil {
		return err
	}

	labels, err := getLabelsFlag(cmd)
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	file, err := openArtifactFile(artifactFile)
	if err != nil {
		return err
	}
	defer file.Close()

	outputArtifact, err := client.CreateArtifact(c.createContext(), environment, newSdkArtifact(name, labels), file)
	if err != nil {
		return err
	}

	return printArtifactOnPrem(cmd, outputArtifact)
}
