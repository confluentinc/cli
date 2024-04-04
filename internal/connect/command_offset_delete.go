package connect

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/retry"
)

func (c *offsetCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a connector's offsets.",
		Args:              cobra.ExactArgs(1),
		RunE:              c.delete,
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete offsets for a connector in the current or specified Kafka cluster context. The behaviour is identical to creating a fresh new connector with the current configs.",
				Code: "confluent connect offset delete lcc-123456",
			},
			examples.Example{
				Code: "confluent connect offset update lcc-123456 --cluster lkc-123456",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *offsetCommand) delete(cmd *cobra.Command, args []string) error {
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
	request := connectv1.ConnectV1AlterOffsetRequest{
		Type: connectv1.ConnectV1AlterOffsetRequestType("DELETE"),
	}

	if err != nil {
		return err
	}

	alterOffsetRequestInfo, err := c.V2Client.AlterConnectorOffsets(connectorName, environmentId, kafkaCluster.ID, request)
	if err != nil {
		return err
	}

	offsetStatus := connectv1.ConnectV1AlterOffsetStatus{
		Request: alterOffsetRequestInfo,
		Status: connectv1.ConnectV1AlterOffsetStatusStatus{
			Phase: "PENDING",
		},
	}

	var apiErr error
	_ = retry.Retry(time.Second, 30*time.Second, func() error {
		offsetStatus, apiErr = c.V2Client.AlterConnectorOffsetsRequestStatus(connectorName, environmentId, kafkaCluster.ID)
		if apiErr != nil {
			return nil
		}

		if strings.ToUpper(offsetStatus.Status.Phase) != "PENDING" {
			return nil
		}
		return fmt.Errorf("delete offset request still pending, checking status again")
	})
	if apiErr != nil {
		return apiErr
	}

	if strings.ToUpper(offsetStatus.Status.Phase) == "PENDING" {
		output.Println(false, "Operation is PENDING. Please run `confluent connect offset status describe` command to get the latest status of the delete request.")
		return nil
	}

	if output.GetFormat(cmd) == output.Human {
		return printHumanDescribeOffsetStatus(cmd, offsetStatus, args[0])
	}

	return printSerializedDescribeOffsetStatus(cmd, offsetStatus, args[0])
}
