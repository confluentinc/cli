package connect

import (
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
		RunE:        c.update,
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

	connectorExpansion, err := c.V2Client.GetConnectorExpansionById(args[0], c.EnvironmentId(), kafkaCluster.ID)
	if err != nil {
		return err
	}

	if _, httpResp, err := c.V2Client.CreateOrUpdateConnectorConfig(connectorExpansion.Info.GetName(), c.EnvironmentId(), kafkaCluster.ID, *userConfigs); err != nil {
		return errors.CatchRequestNotValidMessageError(err, httpResp)
	}

	utils.Printf(cmd, errors.UpdatedConnectorMsg, args[0])
	return nil
}
