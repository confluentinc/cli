package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newArtifactDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe <name>",
		Short:       "Describe a Flink artifact in Confluent Platform.",
		Long:        "Describe a Flink artifact in Confluent Platform. Details reflect the latest version of the artifact.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.artifactDescribeOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) artifactDescribeOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkArtifact, err := client.DescribeArtifact(c.createContext(), environment, name, "")
	if err != nil {
		return err
	}

	return printArtifactOnPrem(cmd, sdkArtifact)
}
