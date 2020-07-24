package auditlog

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
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
			"Use `--combine` to read in multiple Kafka broker server.properties files, " +
			"combine the values of each of their `confluent.security.event.router.config` properties, " +
			"and output a combined configuration to standard output, suitable for centralized audit log " +
			"management, along with any warnings to standard error.",
		RunE: c.config,
		Args: cobra.NoArgs,
	}
	configCmd.Flags().StringToString("combine", nil, `A comma-separated list of k=v pairs, where keys are Kafka cluster IDs, and values are the path to a copy of that cluster's server.properties file. (See https://docs.confluent.io/current/security/rbac/rbac-get-cluster-ids.html for information on how to find your clusters' IDs.) Example: --combine "clusterA=/tmp/cluster/server.properties,clusterB=/tmp/cluster/server.properties".`)
	configCmd.Flags().StringArray("bootstrap-servers", nil, `A public hostname:port of a broker in the Kafka cluster that will receive audit log events. This argument may be repeated multiple times. Example: --bootstrap-servers logs.example.com:9092 --bootstrap-servers logs.example.com:9093.`)
	configCmd.Flags().String("authority", "", `The CRN authority to use in all route patterns. Example: --authority mds.example.com.`)
	configCmd.Flags().SortFlags = false
	c.AddCommand(configCmd)
}

func (c *migrateCmd) config(cmd *cobra.Command, _ []string) error {
	var err error

	crnAuthority := ""
	if cmd.Flags().Changed("authority") {
		crnAuthority, err = cmd.Flags().GetString("authority")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
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
			fileContents, err := ioutil.ReadFile(filePath)
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}
			clusterConfigs[clusterId] = string(fileContents)
		}
	}

	combinedSpec, warnings, err := AuditLogConfigTranslation(clusterConfigs, bootstrapServers, crnAuthority)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	for _, warning := range warnings {
		fmt.Println(warning)
	}

	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	if err = enc.Encode(combinedSpec); err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}
