package connect

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type serializedOffsetStatusDescribeOut struct {
	Id               string           `json:"id" yaml:"id"`
	RequestType      string           `json:"request_type" yaml:"request_type"`
	RequestedAt      string           `json:"requested_at" yaml:"requested_at"`
	Phase            string           `json:"phase" yaml:"phase"`
	Message          string           `json:"message,omitempty" yaml:"message,omitempty"`
	AppliedAt        string           `json:"applied_at,omitempty" yaml:"applied_at,omitempty"`
	RequestedOffsets []map[string]any `json:"requested_offsets,omitempty" yaml:"requested_offsets,omitempty"`
	PreviousOffsets  []map[string]any `json:"previous_offsets,omitempty" yaml:"previous_offsets,omitempty"`
}

type humanOffsetStatusDescribeOut struct {
	Id               string `human:"ID"`
	RequestType      string `human:"RequestType"`
	RequestedAt      string `human:"RequestedAt"`
	Phase            string `human:"Phase"`
	Message          string `human:"Message,omitempty"`
	AppliedAt        string `human:"Applied At,omitempty"`
	RequestedOffsets string `human:"Requested Offsets,omitempty"`
	PreviousOffsets  string `human:"Previous Offsets,omitempty"`
}

func (c *offsetCommand) newStatusDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe connector offset update or delete status.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.statusDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the status of the latest offset update/delete operation for a connector in the current or specified Kafka cluster context.",
				Code: "confluent connect offset status describe lcc-123456",
			},
			examples.Example{
				Code: "confluent connect offset status describe lcc-123456 --cluster lkc-123456",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *offsetCommand) statusDescribe(cmd *cobra.Command, args []string) error {
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

	offsetStatus, err := c.V2Client.AlterConnectorOffsetsRequestStatus(connectorName, environmentId, kafkaCluster.ID)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		return printHumanDescribeOffsetStatus(cmd, offsetStatus, args[0])
	}

	return printSerializedDescribeOffsetStatus(cmd, offsetStatus, args[0])
}

func printHumanDescribeOffsetStatus(cmd *cobra.Command, offsetStatus connectv1.ConnectV1AlterOffsetStatus, id string) error {
	var appliedAt string
	if offsetStatus.AppliedAt.Get() != nil {
		appliedAt = offsetStatus.AppliedAt.Get().String()
	}

	var phase string
	var message string
	_, isStatusSet := offsetStatus.GetStatusOk()
	if isStatusSet {
		phase = offsetStatus.GetStatus().Phase
		if messagePtr := offsetStatus.GetStatus().Message; messagePtr != nil {
			message = *messagePtr
		}
	}

	var prevOffsets []map[string]any
	var offsetStr string
	if offsetStatus.HasPreviousOffsets() {
		prevOffsets = *offsetStatus.PreviousOffsets
		offSetBytes, err := json.Marshal(prevOffsets)
		if err != nil {
			return err
		}
		offsetStr = strings.TrimSpace(string(pretty.Pretty(offSetBytes)))
	}

	var reqOffsets []map[string]any
	var reqOffsetStr string
	if offsetStatus.Request.HasOffsets() {
		reqOffsets = offsetStatus.Request.GetOffsets()
		offSetBytes, err := json.Marshal(reqOffsets)
		if err != nil {
			return err
		}
		reqOffsetStr = strings.TrimSpace(string(pretty.Pretty(offSetBytes)))
	}

	table := output.NewTable(cmd)
	table.Add(&humanOffsetStatusDescribeOut{
		Id:               id,
		RequestType:      string(offsetStatus.Request.Type),
		RequestedAt:      offsetStatus.Request.GetRequestedAt().String(),
		Phase:            phase,
		Message:          message,
		AppliedAt:        appliedAt,
		PreviousOffsets:  offsetStr,
		RequestedOffsets: reqOffsetStr,
	})

	return table.PrintWithAutoWrap(false)
}

func printSerializedDescribeOffsetStatus(cmd *cobra.Command, offsetStatus connectv1.ConnectV1AlterOffsetStatus, id string) error {
	var appliedAt string
	if offsetStatus.AppliedAt.Get() != nil {
		appliedAt = offsetStatus.AppliedAt.Get().String()
	}

	var phase string
	var message string
	_, isStatusSet := offsetStatus.GetStatusOk()
	if isStatusSet {
		phase = offsetStatus.GetStatus().Phase
		if messagePtr := offsetStatus.GetStatus().Message; messagePtr != nil {
			message = *messagePtr
		}
	}

	out := &serializedOffsetStatusDescribeOut{
		Id:               id,
		RequestType:      string(offsetStatus.Request.Type),
		RequestedAt:      offsetStatus.Request.GetRequestedAt().String(),
		Phase:            phase,
		Message:          message,
		AppliedAt:        appliedAt,
		PreviousOffsets:  *offsetStatus.PreviousOffsets,
		RequestedOffsets: offsetStatus.Request.GetOffsets(),
	}

	return output.SerializedOutput(cmd, out)
}
