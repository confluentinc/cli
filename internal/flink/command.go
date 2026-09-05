package flink

import (
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
	cmd.AddCommand(c.newKubernetesClusterCommand())
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
