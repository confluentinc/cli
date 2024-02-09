package flink

import (
	"github.com/spf13/cobra"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type flinkArtifactOut struct {
	Name          string `human:"Name" serialized:"name" `
	PluginId      string `human:"Plugin ID" serialized:"plugin_id" `
	VersionId     string `human:"Version ID" serialized:"version_id" `
	ContentFormat string `human:"Content Format" serialized:"content_format" `
}

func (c *command) newArtifactCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "artifact",
		Short:       "Manage Flink UDF artifact",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}

func printTable(cmd *cobra.Command, plugin connectcustompluginv1.ConnectV1CustomConnectorPlugin) error {
	table := output.NewTable(cmd)
	// Flink UDF artfact lifecycle management is reuse custom connector plugin api for EA customer only
	// Version ID will be surfaced by `ConnectorClass` for EA
	// For Flink GA, Flink team will have public documented API for this. tracking by internal ticket FRT-334
	table.Add(&flinkArtifactOut{
		Name:          plugin.GetDisplayName(),
		PluginId:      plugin.GetId(),
		VersionId:     plugin.GetConnectorClass(),
		ContentFormat: plugin.GetContentFormat(),
	})

	return table.Print()
}
