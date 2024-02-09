package flink

import (
	"github.com/spf13/cobra"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type flinkArtifactSerializedOutOut struct {
	Name           string `serialized:"name"`
	Id             string `serialized:"plugin_id"`
	ConnectorClass string `serialized:"version_id"`
	ContentFormat  string `serialized:"content_format"`
}

// TODO: https://confluentinc.atlassian.net/browse/FRT-334
// This is reuse existing custom connector plugin api for flink udf management for EA customer only
//
//	aka, `ConnectorClass` will return version ID for EA
//	For flink GA, flink team will have public API to do so
type flinkArtifactHumanOut struct {
	Name           string `human:"Name"`
	Id             string `human:"Plugin ID"`
	ConnectorClass string `human:"Version ID"`
	ContentFormat  string `human:"Content Format"`
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
	if output.GetFormat(cmd) == output.Human {
		table.Add(&flinkArtifactHumanOut{
			Id:             plugin.GetId(),
			Name:           plugin.GetDisplayName(),
			ConnectorClass: plugin.GetConnectorClass(),
			ContentFormat:  plugin.GetContentFormat(),
		})
	} else {
		table.Add(&flinkArtifactSerializedOutOut{
			Id:             plugin.GetId(),
			Name:           plugin.GetDisplayName(),
			ConnectorClass: plugin.GetConnectorClass(),
			ContentFormat:  plugin.GetContentFormat(),
		})
	}

	return table.Print()
}
