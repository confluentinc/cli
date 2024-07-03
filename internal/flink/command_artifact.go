package flink

import (
	"github.com/spf13/cobra"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type flinkArtifactOut struct {
	Name          string `human:"Name" serialized:"name" `
	Plugin        string `human:"Plugin" serialized:"plugin" `
	Version       string `human:"Version" serialized:"version" `
	ContentFormat string `human:"Content Format" serialized:"content_format" `
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

func printTable(cmd *cobra.Command, plugin connectcustompluginv1.ConnectV1CustomConnectorPlugin) error {
	table := output.NewTable(cmd)
	table.Add(&flinkArtifactOut{
		Name:          plugin.GetDisplayName(),
		Plugin:        plugin.GetId(),
		Version:       plugin.GetConnectorClass(),
		ContentFormat: plugin.GetContentFormat(),
	})

	return table.Print()
}
