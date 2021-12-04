package connect

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
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
		RunE:        pcmd.NewCLIRunE(c.create),
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
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)

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

	connectorConfig := &schedv1.ConnectorConfig{
		UserConfigs:    *userConfigs,
		AccountId:      c.EnvironmentId(),
		KafkaClusterId: kafkaCluster.ID,
		Name:           (*userConfigs)["name"],
		Plugin:         (*userConfigs)["connector.class"],
	}

	connectorInfo, err := c.Client.Connect.Create(context.Background(), connectorConfig)
	if err != nil {
		return err
	}

	connector := &schedv1.Connector{
		AccountId:      c.EnvironmentId(),
		KafkaClusterId: kafkaCluster.ID,
		Name:           connectorInfo.Name,
	}

	// Resolve Connector ID from name of created connector
	connectorExpansion, err := c.Client.Connect.GetExpansionByName(context.Background(), connector)
	if err != nil {
		return err
	}

	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	trace := connectorExpansion.Status.Connector.Trace
	if outputFormat == output.Human.String() {
		utils.Printf(cmd, errors.CreatedConnectorMsg, connectorInfo.Name, connectorExpansion.Id.Id)
		if trace != "" {
			utils.Printf(cmd, "Error Trace: %s\n", trace)
		}
	} else {
		return output.StructuredOutput(outputFormat, &struct {
			ConnectorName string `json:"name" yaml:"name"`
			Id            string `json:"id" yaml:"id"`
			Trace         string `json:"error_trace,omitempty" yaml:"error_trace,omitempty"`
		}{
			ConnectorName: connectorInfo.Name,
			Id:            connectorExpansion.Id.Id,
			Trace:         trace,
		})
	}

	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, connectorExpansion.Id.Id)
	return nil
}
