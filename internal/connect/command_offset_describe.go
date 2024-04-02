package connect

import (
	"encoding/json"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type serializedOffsetConnectorOut struct {
	Id       string                                      `json:"id" yaml:"id"`
	Name     string                                      `json:"name" yaml:"name"`
	Offsets  []map[string]interface{}                    `json:"offsets,omitempty" yaml:"offsets,omitempty"`
	Metadata connectv1.ConnectV1ConnectorOffsetsMetadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type ConnectDescribeOut struct {
	Id   string `human:"Id"`
	Name string `human:"Name"`
}

type offsetsDescribeOut struct {
	Partition string `human:"Partition"`
	Offset    string `human:"Offset"`
}

type metadataDescribeOut struct {
	ObservedAt string `human:"Observed At"`
}

func (c *offsetCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "get <id>",
		Short:             "Get connector offsets",
		Args:              cobra.ExactArgs(1),
		RunE:              c.getOffset,
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Get connector offsets for a lccId in the current or specified Kafka cluster context.",
				Code: "confluent connect offset get lcc-123456",
			},
			examples.Example{
				Code: "confluent connect offset get lcc-123456 --cluster lkc-123456",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *offsetCommand) getOffset(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	connectorIdToName, err := c.mapConnectorIdToName(environmentId, kafkaCluster.ID)
	if err != nil {
		return err
	}

	offsets, err := c.V2Client.GetConnectorOffset(connectorIdToName[args[0]], environmentId, kafkaCluster.ID)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		return printHumanDescribeOffset(cmd, offsets, args[0], connectorIdToName[args[0]])
	}

	return printSerializedDescribeOffsets(cmd, offsets, args[0], connectorIdToName[args[0]])
}

func (c *offsetCommand) mapConnectorIdToName(environmentId, kafkaClusterId string) (map[string]string, error) {
	// NOTE: Do NOT replace this with `V2Client.GetConnectorExpansionById` calls; that function itself calls `V2Client.ListConnectorsWithExpansions`
	connectors, err := c.V2Client.ListConnectorsWithExpansions(environmentId, kafkaClusterId, "id,info,status")
	if err != nil {
		return nil, err
	}

	connectorIdToName := make(map[string]string)
	for _, connector := range connectors {
		connectorIdToName[connector.Id.GetId()] = connector.Info.GetName()
	}

	return connectorIdToName, nil
}

func printHumanDescribeOffset(cmd *cobra.Command, offsets connectv1.ConnectV1ConnectorOffsets, id string, name string) error {
	output.Println(false, "Connector Details")
	table := output.NewTable(cmd)
	table.Add(&ConnectDescribeOut{
		Name: name,
		Id:   id,
	})
	if err := table.Print(); err != nil {
		return err
	}
	output.Println(false, "")
	output.Println(false, "")

	list := output.NewList(cmd)
	output.Println(false, "Offset Details")
	if &offsets != nil && offsets.HasOffsets() {
		for _, values := range *offsets.Offsets {
			partition := values["partition"]
			partitionString, err := json.Marshal(partition)
			if err != nil {
				return err
			}
			offset := values["offset"]
			offsetString, err := json.Marshal(offset)
			if err != nil {
				return err
			}
			list.Add(&offsetsDescribeOut{
				Partition: string(partitionString),
				Offset:    string(offsetString),
			})
		}
		if err := list.Print(); err != nil {
			return err
		}
		output.Println(false, "")
		output.Println(false, "")
	}

	if &offsets != nil && offsets.HasMetadata() {
		output.Println(false, "Metadata Details")
		list = output.NewList(cmd)

		var observedAt string
		if offsets.Metadata.ObservedAt != nil {
			observedAt = offsets.Metadata.ObservedAt.String()
		}
		list.Add(&metadataDescribeOut{
			ObservedAt: observedAt,
		})
	}

	return list.Print()
}

func printSerializedDescribeOffsets(cmd *cobra.Command, offsets connectv1.ConnectV1ConnectorOffsets, id string, name string) error {

	var offsetInfo []map[string]interface{}
	var metadata connectv1.ConnectV1ConnectorOffsetsMetadata
	if &offsets != nil && offsets.HasMetadata() {
		metadata = *offsets.Metadata
	}

	if &offsets != nil && offsets.HasOffsets() {
		offsetInfo = *offsets.Offsets
	}

	out := &serializedOffsetConnectorOut{
		Id:       id,
		Name:     name,
		Offsets:  offsetInfo,
		Metadata: metadata,
	}
	return output.SerializedOutput(cmd, out)
}
