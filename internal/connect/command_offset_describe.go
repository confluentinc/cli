package connect

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/retry"
)

type humanOffsetConnectorOut struct {
	Id       string `human:"ID"`
	Name     string `human:"Name"`
	Offsets  string `human:"Offsets"`
	Metadata string `human:"Metadata"`
}

type serializedOffsetConnectorOut struct {
	Id       string                                      `json:"id" yaml:"id"`
	Name     string                                      `json:"name" yaml:"name"`
	Offsets  []map[string]any                            `json:"offsets" yaml:"offsets"`
	Metadata connectv1.ConnectV1ConnectorOffsetsMetadata `json:"metadata" yaml:"metadata"`
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

	cmd.Flags().Int32("staleness-threshold", 120, "Repeatedly fetch offsets, until receiving an offset with an observed time within the staleness threshold in seconds, for a minimum of 5 seconds.")
	cmd.Flags().Int32("timeout", 30, "Max time in seconds to wait until we get an offset within the staleness threshold.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *offsetCommand) describe(cmd *cobra.Command, args []string) error {
	stalenessThreshold, err := cmd.Flags().GetInt32("staleness-threshold")
	if err != nil {
		return err
	}

	if stalenessThreshold < 5 {
		return fmt.Errorf("`--staleness-threshold` cannot be less than 5 seconds")
	}

	timeout, err := cmd.Flags().GetInt32("timeout")
	if err != nil {
		return err
	}
	if timeout <= 0 {
		return fmt.Errorf("`--timeout` has to be a positive value")
	}

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
	var apiErr error
	var offsets connectv1.ConnectV1ConnectorOffsets
	_ = retry.Retry(time.Second, time.Duration(timeout)*time.Second, func() error {
		offsets, apiErr = c.V2Client.GetConnectorOffset(connector.Info.GetName(), environmentId, kafkaCluster.ID)
		if apiErr != nil {
			return apiErr
		}

		if offsets.HasMetadata() && offsets.Metadata.HasObservedAt() && int32(time.Since(*offsets.Metadata.ObservedAt).Seconds()) <= stalenessThreshold {
			return nil
		}

		return fmt.Errorf("got stale offsets, fetching again")
	})
	if apiErr != nil {
		return apiErr
	}

	if output.GetFormat(cmd) == output.Human {
		return printHumanDescribeOffset(cmd, offsets, args[0], connectorName)
	}

	return printSerializedDescribeOffsets(cmd, offsets, args[0], connectorName)
}

func printHumanDescribeOffset(cmd *cobra.Command, offsets connectv1.ConnectV1ConnectorOffsets, id, name string) error {
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

	var offsetInfo []map[string]any
	var offsetStr string
	if offsets.HasOffsets() {
		offsetInfo = *offsets.Offsets
		offSetBytes, err := json.Marshal(offsetInfo)
		if err != nil {
			return err
		}
		offsetStr = string(pretty.Pretty(offSetBytes))
	}

	table := output.NewTable(cmd)

	table.Add(&humanOffsetConnectorOut{
		Id:       id,
		Name:     name,
		Offsets:  offsetStr,
		Metadata: metadataStr,
	})

	return table.PrintWithAutoWrap(false)
}

func printSerializedDescribeOffsets(cmd *cobra.Command, offsets connectv1.ConnectV1ConnectorOffsets, id, name string) error {
	var metadata connectv1.ConnectV1ConnectorOffsetsMetadata
	if offsets.HasMetadata() {
		metadata = *offsets.Metadata
	}

	var offsetInfo []map[string]any
	if offsets.HasOffsets() {
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
