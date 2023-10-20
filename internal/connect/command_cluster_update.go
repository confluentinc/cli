package connect

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/properties"
	"github.com/confluentinc/cli/v3/pkg/resource"
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
	cmd.Flags().String("config-file", "", "JSON connector configuration file.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "json"))

	return cmd
}

func (c *clusterCommand) update(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(c.V2Client)
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	var userConfigs *map[string]string
	if cmd.Flags().Changed("config") {
		configs, err := cmd.Flags().GetStringSlice("config")
		if err != nil {
			return err
		}
		configMap, err := properties.ConfigFlagToMap(configs)
		if err != nil {
			return err
		}

		connector, err := c.V2Client.GetConnectorExpansionById(args[0], environmentId, kafkaCluster.ID)
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
		return fmt.Errorf("one of `--config` or `--config-file` must be specified")
	}

	connector, err := c.V2Client.GetConnectorExpansionById(args[0], environmentId, kafkaCluster.ID)
	if err != nil {
		return err
	}

	if _, err := c.V2Client.CreateOrUpdateConnectorConfig(connector.Info.GetName(), environmentId, kafkaCluster.ID, *userConfigs); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UpdatedResourceMsg, resource.Connector, args[0])
	return nil
}
