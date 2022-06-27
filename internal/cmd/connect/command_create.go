package connect

import (
	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "create",
		Short:       "Create a connector.",
		Args:        cobra.NoArgs,
		RunE:        c.create,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a connector in the current or specified Kafka cluster context.",
				Code: "confluent connect create --config config.json",
			},
			examples.Example{
				Code: "confluent connect create --config config.json --cluster lkc-123456",
			},
		),
	}

	cmd.Flags().String("config", "", "JSON connector config file.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("config")

	return cmd
}

func (c *command) create(cmd *cobra.Command, _ []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	userConfigs, err := getConfig(cmd)
	if err != nil {
		return err
	}

	connectConfig := connectv1.InlineObject{
		Name:   connectv1.PtrString((*userConfigs)["name"]),
		Config: userConfigs,
	}

	connectorInfo, httpResp, err := c.V2Client.CreateConnector(c.EnvironmentId(), kafkaCluster.ID, connectConfig)
	if err != nil {
		return errors.CatchRequestNotValidMessageError(err, httpResp)
	}

	connectorExpansion, err := c.V2Client.GetConnectorExpansionByName(connectorInfo.GetName(), c.EnvironmentId(), kafkaCluster.ID)
	if err != nil {
		return err
	}

	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	trace := connectorExpansion.Status.Connector.GetTrace()

	if outputFormat == output.Human.String() {
		utils.Printf(cmd, errors.CreatedConnectorMsg, connectorInfo.GetName(), connectorExpansion.Id.GetId())
		if trace != "" {
			utils.Printf(cmd, "Error Trace: %s\n", trace)
		}
	} else {
		return output.StructuredOutput(outputFormat, &connectCreateDisplay{
			Id:         connectorExpansion.Id.GetId(),
			Name:       connectorInfo.GetName(),
			ErrorTrace: trace,
		})
	}

	return nil
}
