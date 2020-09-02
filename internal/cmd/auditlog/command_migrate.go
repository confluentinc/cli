package auditlog

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
	"os"
)

type migrateCmd struct {
	*cmd.CLICommand
	prerunner cmd.PreRunner
}

func NewMigrateCommand(prerunner cmd.PreRunner) *cobra.Command {
	cliCmd := cmd.NewCLICommand(
		&cobra.Command{
			Use:   "migrate",
			Short: "Migrate legacy audit log configurations.",
		}, prerunner)
	command := &migrateCmd{
		CLICommand: cliCmd,
		prerunner:  prerunner,
	}
	command.init()
	return command.Command
}

func (c *migrateCmd) init() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Migrate legacy audit log configurations.",
		Long: "Migrate legacy audit log configurations. " +
			"Use ``--combine`` to read in multiple Kafka broker ``server.properties`` files, " +
			"combine the values of their ``confluent.security.event.router.config`` properties, " +
			"and output a combined configuration suitable for centralized audit log " +
			"management. This is sent to standard output along with any warnings to standard error.",
		RunE: c.migrate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Combine two audit log configuration files for clusters 'clusterA' and 'clusterB' with the following bootstrap servers and authority.",
				Code: "confluent audit-log migrate config --combine clusterA=/tmp/cluster/server.properties,clusterB=/tmp/cluster/server.properties " +
					"--bootstrap-servers logs.example.com:9092 --bootstrap-servers logs.example.com:9093 --authority mds.example.com",
			},
		),
		Args: cobra.NoArgs,
	}
	configCmd.Flags().StringToString("combine", nil, `A comma-separated list of k=v pairs, where keys are Kafka cluster IDs, and values are the path to that cluster's server.properties file.`)
	configCmd.Flags().StringArray("bootstrap-servers", nil, `A public hostname:port of a broker in the Kafka cluster that will receive audit log events.`)
	configCmd.Flags().String("authority", "", `The CRN authority to use in all route patterns.`)
	configCmd.Flags().SortFlags = false
	c.AddCommand(configCmd)
}

func (c *migrateCmd) migrate(cmd *cobra.Command, _ []string) error {
	var err error

	crnAuthority, err := cmd.Flags().GetString("authority")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	bootstrapServers := []string{}
	if cmd.Flags().Changed("bootstrap-servers") {
		bootstrapServers, err = cmd.Flags().GetStringArray("bootstrap-servers")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	}

	clusterConfigs := map[string]string{}
	if cmd.Flags().Changed("combine") {
		fileNameMap, err := cmd.Flags().GetStringToString("combine")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}

		for clusterId, filePath := range fileNameMap {
			propertyFile, err := utils.LoadPropertiesFile(filePath)
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}

			routerConfig, ok := propertyFile.Get("confluent.security.event.router.config")
			if !ok {
				fmt.Println(fmt.Sprintf("Ignoring property file %s because it does not contain a router configuration.", filePath))
				continue
			}
			clusterConfigs[clusterId] = routerConfig
		}
	}

	combinedSpec, warnings, err := AuditLogConfigTranslation(clusterConfigs, bootstrapServers, crnAuthority)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, warning)
		fmt.Println()
	}

	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	if err = enc.Encode(combinedSpec); err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}
