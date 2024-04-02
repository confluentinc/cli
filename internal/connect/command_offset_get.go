package connect

import (
	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type serializedOffsetOut struct {
	Connector *serializedOffsetConnectorOut `json:"connector" yaml:"connector"`
	Tasks     []serializedTasksOut          `json:"tasks" yaml:"tasks"`
	Configs   []serializedConfigsOut        `json:"configs" yaml:"configs"`
}

type serializedOffsetConnectorOut struct {
	Id       string                                      `json:"id" yaml:"id"`
	Name     string                                      `json:"name" yaml:"name"`
	Offsets  []map[string]interface{}                    `json:"offsets,omitempty" yaml:"offsets"`
	Metadata connectv1.ConnectV1ConnectorOffsetsMetadata `json:"metadata,omitempty" yaml:"metadata"`
}

type offsetsDescribeOut struct {
	Partition interface{} `human:"Partition"`
	Offset    interface{} `human:"Offset"`
}

type metadataDescribeOut struct {
	ObservedAt string `human:"Observed At"`
}

func (c *offsetCommand) newGetOffsetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "get <id>",
		Short:       "Get connector offsets",
		Args:        cobra.ExactArgs(1),
		RunE:        c.getOffset,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
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
		return printHumanDescribeOffset(cmd, offsets)
	}

	return printSerializedDescribeOffsets(cmd, offsets)
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

func printHumanDescribeOffset(cmd *cobra.Command, offset connectv1.ConnectV1ConnectorOffsets) error {
	output.Println(false, "Connector Details")
	table := output.NewTable(cmd)
	table.Add(&connectOut{
		Name: *offset.Name,
		Id:   *offset.Id,
	})
	if err := table.Print(); err != nil {
		return err
	}
	output.Println(false, "")
	output.Println(false, "")

	output.Println(false, "Offset Details")
	list := output.NewList(cmd)
	for _, values := range *offset.Offsets {
		list.Add(&offsetsDescribeOut{
			Partition: values["partition"],
			Offset:    values["offset"],
		})
	}
	if err := list.Print(); err != nil {
		return err
	}
	output.Println(false, "")
	output.Println(false, "")

	output.Println(false, "Metadata Details")
	list = output.NewList(cmd)

	list.Add(&metadataDescribeOut{
		ObservedAt: offset.Metadata.ObservedAt.String(),
	})

	return list.Print()
}

func printSerializedDescribeOffsets(cmd *cobra.Command, offsets connectv1.ConnectV1ConnectorOffsets) error {

	out := &serializedOffsetConnectorOut{
		Id:       *offsets.Id,
		Name:     *offsets.Name,
		Offsets:  *offsets.Offsets,
		Metadata: *offsets.Metadata,
	}
	return output.SerializedOutput(cmd, out)
}
