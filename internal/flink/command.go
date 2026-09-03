package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flink",
		Short: "Manage Apache Flink.",
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	// On-prem commands are able to run with or without login. Accordingly, set the pre-runner.
	if cfg.IsOnPremLogin() {
		c = &command{pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)}
	} else if !cfg.IsCloudLogin() {
		cmd.PersistentPreRunE = prerunner.Anonymous(c.AuthenticatedCLICommand.CLICommand, false)
	}

	// On-Prem Specific Commands
	cmd.AddCommand(c.newApplicationCommand())
	cmd.AddCommand(c.newCatalogCommand())
	cmd.AddCommand(c.newDetachedSavepointCommand())
	cmd.AddCommand(c.newEnvironmentCommand())
	cmd.AddCommand(c.newSavepointCommand())
	cmd.AddCommand(c.newSecretCommand())
	cmd.AddCommand(c.newSecretMappingCommand())
	cmd.AddCommand(c.newSystemInfoCommand())

	// On-Prem and Cloud Shared Commands
	if cfg.IsCloudLogin() {
		cmd.AddCommand(
			newComputePoolCommand(cfg, prerunner),
			newStatementCommand(cfg, prerunner),
		)
	} else {
		cmd.AddCommand(c.newComputePoolCommandOnPrem())
		cmd.AddCommand(c.newStatementCommandOnPrem())
	}

	if !cfg.IsOnPremLogin() {
		cmd.AddCommand(c.newShellCommand(prerunner, cfg))
	}

	// Cloud Specific Commands
	cmd.AddCommand(c.newArtifactCommand())
	cmd.AddCommand(c.newConnectionCommand())
	cmd.AddCommand(c.newConnectivityTypeCommand())
	cmd.AddCommand(c.newEndpointCommand())
	cmd.AddCommand(c.newMaterializedTableCommand())

	// Generated Cloud commands
	cmd.AddCommand(
		newOrgComputePoolConfigCommand(cfg, prerunner),
		newRegionCommand(cfg, prerunner),
		// cli-tfgen:cli-subcommands
	)

	return cmd
}

func (c *command) addDatabaseFlag(cmd *cobra.Command) {
	cmd.Flags().String("database", "", "The database which will be used as the default database. When using Kafka, this is the cluster ID.")
	pcmd.RegisterFlagCompletionFunc(cmd, "database", c.autocompleteDatabases)
}

func (c *command) autocompleteDatabases(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	clusters, err := c.V2Client.ListKafkaClusters(environmentId)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(clusters))
	for i, cluster := range clusters {
		suggestions[i] = fmt.Sprintf("%s\t%s", cluster.GetId(), cluster.Spec.GetDisplayName())
	}
	return suggestions
}
