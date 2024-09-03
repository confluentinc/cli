package flink

import (
	"github.com/spf13/cobra"

	flinkartifactv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-artifact/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type flinkArtifactOut struct {
	Id            string `human:"ID" serialized:"id"`
	Name          string `human:"Name" serialized:"name"`
	Version       string `human:"Version" serialized:"version"`
	Class         string `human:"Class" serialized:"class"`
	Cloud         string `human:"Cloud" serialized:"cloud"`
	Region        string `human:"Region" serialized:"region"`
	Environment   string `human:"Environment" serialized:"environment"`
	ContentFormat string `human:"Content Format" serialized:"content_format"`
}

func (c *command) newArtifactCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "artifact",
		Short:       "Manage Flink UDF artifacts.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}

func printTable(cmd *cobra.Command, plugin flinkartifactv1.ArtifactV1FlinkArtifact) error {
	table := output.NewTable(cmd)

	var pluginVersion = ""
	if len(plugin.GetVersions()) > 0 {
		pluginVersion = (*plugin.Versions)[0].GetVersion()
	}

	table.Add(&flinkArtifactOut{
		Name:          plugin.GetDisplayName(),
		Id:            plugin.GetId(),
		Version:       pluginVersion,
		Class:         plugin.GetClass(),
		Cloud:         plugin.GetCloud(),
		Region:        plugin.GetRegion(),
		Environment:   plugin.GetEnvironment(),
		ContentFormat: plugin.GetContentFormat(),
	})

	return table.Print()
}
