package connect

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/retry"
)

func (c *offsetCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a connector's offsets.",
		Args:              cobra.ExactArgs(1),
		RunE:              c.update,
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "The configuration file contains offsets to be set for the connector.",
				Code: `{
  "offsets": [
	{
	  "partition": {
		"kafka_partition": 0,
		"kafka_topic": "topic_A"
	  },
	  "offset": {
		"kafka_offset": 1000
	  }
	}
  ]
}`,
			},
			examples.Example{
				Text: "Update offsets for a connector in the current or specified Kafka cluster context.",
				Code: "confluent connect offset update lcc-123456 --config-file config.json",
			},
			examples.Example{
				Code: "confluent connect offset update lcc-123456 --config-file config.json --cluster lkc-123456",
			},
		),
	}

	cmd.Flags().String("config-file", "", "JSON file containing new connector offsets.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "json"))
	cobra.CheckErr(cmd.MarkFlagRequired("config-file"))

	return cmd
}

func (c *offsetCommand) update(cmd *cobra.Command, args []string) error {
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
	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}

	request, err := c.getAlterOffsetRequestBody(configFile)
	if err != nil {
		return err
	}

	alterOffsetRequestInfo, err := c.V2Client.AlterConnectorOffsets(connectorName, environmentId, kafkaCluster.ID, *request)
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

		if strings.ToUpper(offsetStatus.Status.GetPhase()) != "PENDING" {
			return nil
		}
		return fmt.Errorf("update offset request still pending, checking status again")
	})
	if apiErr != nil {
		return apiErr
	}

	if strings.ToUpper(offsetStatus.Status.GetPhase()) == "PENDING" {
		output.Println(c.Config.EnableColor, "Operation is PENDING. Please run `confluent connect offset status describe` to get the latest status of the update request.")
		return nil
	}

	if output.GetFormat(cmd) == output.Human {
		return printHumanDescribeOffsetStatus(cmd, offsetStatus, args[0])
	}

	return printSerializedDescribeOffsetStatus(cmd, offsetStatus, args[0])
}

func (c *offsetCommand) getAlterOffsetRequestBody(configFile string) (*connectv1.ConnectV1AlterOffsetRequest, error) {
	jsonFile, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, configFile, err)
	}

	if len(jsonFile) == 0 {
		return nil, fmt.Errorf(`offset configuration file "%s" is empty`, configFile)
	}

	var request *connectv1.ConnectV1AlterOffsetRequest
	if err := json.Unmarshal(jsonFile, &request); err != nil {
		return nil, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, configFile, err)
	}

	request.SetType("PATCH")
	return request, nil
}
