package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newArtifactListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list",
		Short:       "List Flink artifacts in Confluent Platform.",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.artifactListOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) artifactListOnPrem(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkArtifacts, err := client.ListArtifacts(c.createContext(), environment)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, artifact := range sdkArtifacts {
			list.Add(newArtifactOutOnPrem(artifact))
		}
		return list.Print()
	}

	localArtifacts := make([]LocalArtifact, 0, len(sdkArtifacts))
	for _, sdkArtifact := range sdkArtifacts {
		localArtifacts = append(localArtifacts, convertSdkArtifactToLocalArtifact(sdkArtifact))
	}

	return output.SerializedOutput(cmd, localArtifacts)
}
