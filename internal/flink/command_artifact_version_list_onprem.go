package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newArtifactVersionListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "list <name>",
		Short:       "List the versions of a Flink artifact in Confluent Platform.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.artifactVersionListOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) artifactVersionListOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkVersions, err := client.ListArtifactVersions(c.createContext(), environment, name)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		// Preserve the server's newest-first ordering instead of sorting version strings lexicographically.
		list.Sort(false)
		for _, version := range sdkVersions {
			list.Add(newArtifactVersionOutOnPrem(version))
		}
		return list.Print()
	}

	localVersions := make([]LocalArtifact, 0, len(sdkVersions))
	for _, sdkVersion := range sdkVersions {
		localVersions = append(localVersions, convertSdkArtifactToLocalArtifact(sdkVersion))
	}

	return output.SerializedOutput(cmd, localVersions)
}
