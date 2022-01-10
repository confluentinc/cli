package connect

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "update <id>",
		Short:       "Update a connector configuration.",
		Args:        cobra.ExactArgs(1),
		RunE:        pcmd.NewCLIRunE(c.update),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.Flags().String("config", "", "JSON connector config file.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	_ = cmd.MarkFlagRequired("config")

	return cmd
}

func (c *command) update(cmd *cobra.Command, args []string) error {
	userConfigs, err := getConfig(cmd)
	if err != nil {
		return err
	}

	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	connector := &schedv1.Connector{
		AccountId:      c.EnvironmentId(),
		KafkaClusterId: kafkaCluster.ID,
		Id:             args[0],
	}

	// Resolve Connector Name from ID
	connectorExpansion, err := c.Client.Connect.GetExpansionById(context.Background(), connector)
	if err != nil {
		return err
	}

	connectorConfig := &schedv1.ConnectorConfig{
		UserConfigs:    *userConfigs,
		AccountId:      c.EnvironmentId(),
		KafkaClusterId: kafkaCluster.ID,
		Name:           connectorExpansion.Info.Name,
		Plugin:         (*userConfigs)["connector.class"],
	}

	if _, err := c.Client.Connect.Update(context.Background(), connectorConfig); err != nil {
		return err
	}

	utils.Printf(cmd, errors.UpdatedConnectorMsg, args[0])
	return nil
}
