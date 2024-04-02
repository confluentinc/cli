package connect

import (
	"encoding/json"
	"fmt"
	"os"
	time2 "time"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *offsetCommand) newAlterOffsetCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:         "alter <id>",
		Short:       "Alter a connector's offsets'",
		Args:        cobra.ExactArgs(1),
		RunE:        c.alterOffset,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Alter connector offsets for a lccId in the current or specified Kafka cluster context.",
				Code: "confluent connect offset alter lcc-123456 --config-file config.json",
			},
			examples.Example{
				Code: "confluent connect offset alter lcc-123456 --config-file config.json --cluster lkc-123456",
			},
		),
	}

	cmd.Flags().String("config-file", "", "JSON file containing connector offsets to set to.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "json"))
	cobra.CheckErr(cmd.MarkFlagRequired("config-file"))

	return cmd
}

func (c *offsetCommand) alterOffset(cmd *cobra.Command, args []string) error {
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

	request, err := c.getAlterOffsetRequestBody(cmd)
	if err != nil {
		return err
	}

	_, err = c.V2Client.AlterConnectorOffsets(connectorIdToName[args[0]], environmentId, kafkaCluster.ID, *request)
	if err != nil {
		return err
	}

	var offsetStatus connectv1.ConnectV1AlterOffsetStatus
	currTime := time2.Now()
	table := output.NewTable(cmd)
	for {
		offsetStatus, err = c.V2Client.AlterConnectorOffsetsRequestStatus(connectorIdToName[args[0]], environmentId, kafkaCluster.ID)
		if err != nil {
			return err
		}

		if offsetStatus.GetStatus().Phase != "PENDING" || time2.Since(currTime).Seconds() > 30 {
			table.Add(&alterStatusOut{
				Id:        args[0],
				Phase:     offsetStatus.GetStatus().Phase,
				Message:   *offsetStatus.GetStatus().Message,
				AppliedAt: offsetStatus.AppliedAt.Get().String(),
			})
		}

		return table.Print()
	}
}

func (c *offsetCommand) getAlterOffsetRequestBody(cmd *cobra.Command) (*connectv1.ConnectV1AlterOffsetRequest, error) {

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return nil, err
	}

	jsonFile, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, jsonFile, err)
	}
	if len(jsonFile) == 0 {
		return nil, fmt.Errorf(`alter offset config file "%s" is empty`, jsonFile)
	}

	var request connectv1.ConnectV1AlterOffsetRequest
	if err := json.Unmarshal(jsonFile, &request); err != nil {
		return nil, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, jsonFile, err)
	}

	return &request, err
}
