package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *clusterCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "update <id>",
		Short:       "Update a connector configuration.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.update,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the connector being updated.`)
	cmd.Flags().String("config-file", "", "JSON connector config file.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *clusterCommand) update(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	var userConfigs *map[string]string
	if cmd.Flags().Changed("config") {
		configList, err := cmd.Flags().GetStringSlice("config")
		if err != nil {
			return err
		}
		configMap, err := properties.ConfigFlagToMap(configList)
		if err != nil {
			return err
		}

		connector, err := c.V2Client.GetConnectorExpansionById(args[0], c.EnvironmentId(), kafkaCluster.ID)
		if err != nil {
			return err
		}
		currentConfigs := connector.Info.GetConfig()

		for name, value := range configMap {
			currentConfigs[name] = value
		}
		userConfigs = &currentConfigs
	} else if cmd.Flags().Changed("config-file") {
		userConfigs, err = getConfig(cmd)
		if err != nil {
			return err
		}
	} else {
		return errors.New("one of `--config` or `--config-file` must be specified")
	}

	connector, err := c.V2Client.GetConnectorExpansionById(args[0], c.EnvironmentId(), kafkaCluster.ID)
	if err != nil {
		return err
	}

	if _, err := c.V2Client.CreateOrUpdateConnectorConfig(connector.Info.GetName(), c.EnvironmentId(), kafkaCluster.ID, *userConfigs); err != nil {
		return err
	}

	utils.Printf(cmd, errors.UpdatedResourceMsg, resource.Connector, args[0])
	return nil
}
