package connect

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type OffsetConnectorOut struct {
	Id       string `human:"ID" json:"id" yaml:"id"`
	Name     string `human:"Name" json:"name" yaml:"name"`
	Offsets  string `human:"Offsets" json:"offsets,omitempty" yaml:"offsets,omitempty"`
	Metadata string `human:"Metadata" json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type SerializedOffsetConnectorOut struct {
	Id       string                                      `human:"ID" json:"id" yaml:"id"`
	Name     string                                      `human:"Name" json:"name" yaml:"name"`
	Offsets  []map[string]any                            `human:"Offsets" json:"offsets,omitempty" yaml:"offsets,omitempty"`
	Metadata connectv1.ConnectV1ConnectorOffsetsMetadata `human:"Metadata" json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

func (c *offsetCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe connector offsets.",
		Args:              cobra.ExactArgs(1),
		RunE:              c.describe,
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe offsets for a connector in the current or specified Kafka cluster context.",
				Code: "confluent connect offset describe lcc-123456",
			},
			examples.Example{
				Code: "confluent connect offset describe lcc-123456 --cluster lkc-123456",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *offsetCommand) describe(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	connector, err := c.V2Client.GetConnectorExpansionById(args[0], environmentId, kafkaCluster.ID)
	if err != nil {
		return err
	}

	connectorName := connector.Info.GetName()
	offsets, err := c.V2Client.GetConnectorOffset(connector.Info.GetName(), environmentId, kafkaCluster.ID)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		return printHumanDescribeOffset(cmd, offsets, args[0], connectorName)
	}

	return printSerializedDescribeOffsets(cmd, offsets, args[0], connectorName)
}

func printHumanDescribeOffset(cmd *cobra.Command, offsets connectv1.ConnectV1ConnectorOffsets, id string, name string) error {
	var offsetInfo []map[string]any
	var metadata connectv1.ConnectV1ConnectorOffsetsMetadata
	var metadataStr string
	if offsets.HasMetadata() {
		metadata = *offsets.Metadata
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		metadataStr = string(pretty.Pretty(metadataBytes))
	}

	var offsetStr string
	if offsets.HasOffsets() {
		offsetInfo = *offsets.Offsets
		offSetBytes, err := json.Marshal(offsetInfo)
		if err != nil {
			return err
		}
		offsetStr = string(pretty.Pretty(offSetBytes))
	}

	list := output.NewList(cmd)

	list.Add(&OffsetConnectorOut{
		Id:       id,
		Name:     name,
		Offsets:  offsetStr,
		Metadata: metadataStr,
	})

	return list.PrintWithAutoWrap(false)
}

func printSerializedDescribeOffsets(cmd *cobra.Command, offsets connectv1.ConnectV1ConnectorOffsets, id string, name string) error {
	var offsetInfo []map[string]any
	var metadata connectv1.ConnectV1ConnectorOffsetsMetadata
	if offsets.HasMetadata() {
		metadata = *offsets.Metadata
	}
	if offsets.HasOffsets() {
		offsetInfo = *offsets.Offsets
	}

	out := &SerializedOffsetConnectorOut{
		Id:       id,
		Name:     name,
		Offsets:  offsetInfo,
		Metadata: metadata,
	}
	return output.SerializedOutput(cmd, out)
}
