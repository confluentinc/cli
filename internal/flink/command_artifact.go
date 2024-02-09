package flink

import (
	"github.com/spf13/cobra"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

// TODO: https://confluentinc.atlassian.net/browse/FRT-334
// This is reuse existing custom connector plugin api for flink udf management for EA customer only
//
//	aka, `ConnectorClass` will return version ID for EA
//	For flink GA, flink team will have public API to do so
type flinkArtifactOut struct {
	Name           string `serialized:"name" human:"Name"`
	Id             string `serialized:"plugin_id" human:"Plugin ID"`
	ConnectorClass string `serialized:"version_id" human:"Version ID"`
	ContentFormat  string `serialized:"content_format" human:"Content Format"`
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
	table.Add(&flinkArtifactOut{
		Id:             plugin.GetId(),
		Name:           plugin.GetDisplayName(),
		ConnectorClass: plugin.GetConnectorClass(),
		ContentFormat:  plugin.GetContentFormat(),
	})

	return table.Print()
}
