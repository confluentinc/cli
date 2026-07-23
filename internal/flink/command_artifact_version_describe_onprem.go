package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newArtifactVersionDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe <name>",
		Short:       "Describe a Flink artifact version in Confluent Platform.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.artifactVersionDescribeOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe version 2 of Flink artifact "my-artifact" in the environment "my-environment".`,
				Code: "confluent flink artifact version describe my-artifact --version 2 --environment my-environment",
			},
		),
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("version", "", "Version of the artifact to describe.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("version"))

	return cmd
}

func (c *command) artifactVersionDescribeOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkArtifact, err := client.DescribeArtifact(c.createContext(), environment, name, version)
	if err != nil {
		return err
	}

	return printArtifactVersionOnPrem(cmd, sdkArtifact)
}
