package auditlog

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *configCommand) newMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate legacy audit log configurations.",
		Long: "Migrate legacy audit log configurations. " +
			"Use `--combine` to read in multiple Kafka broker `server.properties` files, " +
			"combine the values of their `confluent.security.event.router.config` properties, " +
			"and output a combined configuration suitable for centralized audit log " +
			"management. This is sent to standard output along with any warnings to standard error.",
		Args: cobra.NoArgs,
		RunE: c.migrate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Combine two audit log configuration files for clusters 'clusterA' and 'clusterB' with the following bootstrap servers and authority.",
				Code: "confluent audit-log config migrate --combine clusterA=/tmp/cluster/server.properties,clusterB=/tmp/cluster/server.properties " +
					"--bootstrap-servers logs.example.com:9092,logs.example.com:9093 --authority mds.example.com",
			},
		),
	}

	cmd.Flags().StringToString("combine", nil, `A comma-separated list of k=v pairs, where keys are Kafka cluster IDs, and values are the path to that cluster's server.properties file.`)
	cmd.Flags().StringSlice("bootstrap-servers", nil, `A comma-separated list of public brokers ("hostname:port") in the Kafka cluster that will receive audit log events.`)
	cmd.Flags().String("authority", "", `The CRN authority to use in all route patterns.`)

	return cmd
}

func (c *configCommand) migrate(cmd *cobra.Command, _ []string) error {
	authority, err := cmd.Flags().GetString("authority")
	if err != nil {
		return err
	}

	bootstrapServers := []string{}
	if cmd.Flags().Changed("bootstrap-servers") {
		bootstrapServers, err = cmd.Flags().GetStringSlice("bootstrap-servers")
		if err != nil {
			return err
		}
	}

	clusterConfigs := map[string]string{}
	if cmd.Flags().Changed("combine") {
		fileNameMap, err := cmd.Flags().GetStringToString("combine")
		if err != nil {
			return err
		}

		for clusterId, filePath := range fileNameMap {
			propertyFile, err := utils.LoadPropertiesFile(filePath)
			if err != nil {
				return err
			}

			routerConfig, ok := propertyFile.Get("confluent.security.event.router.config")
			if !ok {
				fmt.Printf("Ignoring property file %s because it does not contain a router configuration.\n", filePath)
				continue
			}
			clusterConfigs[clusterId] = routerConfig
		}
	}

	combinedSpec, warnings, err := AuditLogConfigTranslation(clusterConfigs, bootstrapServers, authority)
	if err != nil {
		return err
	}
	for _, warning := range warnings {
		fmt.Fprintln(os.Stderr, warning)
		fmt.Println()
	}

	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")

	return enc.Encode(combinedSpec)
}
