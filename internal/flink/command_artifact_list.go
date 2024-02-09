package flink

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type customPluginOutList struct {
	Id   string `human:"ID" serialized:"id"`
	Name string `human:"Name" serialized:"name"`
}

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink UDF artifacts.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List Flink UDF artifacts in the org",
				Code: "confluent flink artifact list",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	plugins, err := c.V2Client.ListCustomPlugins()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, plugin := range plugins {
		if strings.ToLower(plugin.GetConnectorType()) != "flink-udf" {
			continue
		}
		list.Add(&customPluginOutList{
			Name: plugin.GetDisplayName(),
			Id:   plugin.GetId(),
		})
	}
	return list.Print()
}
