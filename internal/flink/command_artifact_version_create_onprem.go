package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newArtifactVersionCreateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Upload a new Flink artifact version in Confluent Platform.",
		Long:  "Upload a new version of a Flink artifact in Confluent Platform. If the uploaded content is identical to the latest version, the artifact is left unchanged and no new version is created.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.artifactVersionCreateOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Upload the file "my-udf-2.0.jar" as a new version of Flink artifact "my-udf.jar" in the environment "my-environment".`,
				Code: "confluent flink artifact version create my-udf.jar --artifact-file my-udf-2.0.jar --environment my-environment",
			},
		),
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("artifact-file", "", "Path to the Flink artifact JAR or ZIP file.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("artifact-file"))
	cobra.CheckErr(cmd.MarkFlagFilename("artifact-file", "jar", "zip"))

	return cmd
}

func (c *command) artifactVersionCreateOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	artifactFile, err := cmd.Flags().GetString("artifact-file")
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

	// Nil labels leave existing artifact labels unchanged; only the content (new version) is uploaded.
	outputArtifact, err := client.UpdateArtifact(c.createContext(), environment, name, newSdkArtifact(name, nil), file)
	if err != nil {
		return err
	}

	// When the upload is identical to the latest version, CMF creates no new version: it returns the latest one and says
	// so in the status message. Serialized output already includes the message, so only the table needs it printed.
	if status := outputArtifact.GetStatus(); status.GetMessage() != "" && output.GetFormat(cmd) == output.Human {
		output.Println(false, status.GetMessage())
	}

	return printArtifactVersionOnPrem(cmd, outputArtifact)
}
