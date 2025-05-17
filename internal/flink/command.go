package flink

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
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
	cmd.AddCommand(c.newEnvironmentCommand())

	// Cloud Specific Commands
	if cfg.IsTest || featureflags.Manager.BoolVariation("cli.flink", cfg.Context(), config.CliLaunchDarklyClient, true, false) {
		if cfg.IsTest || featureflags.Manager.BoolVariation("cli.flink.connection", cfg.Context(), config.CliLaunchDarklyClient, true, false) {
			cmd.AddCommand(c.newConnectionCommand())
		}

		cmd.AddCommand(c.newArtifactCommand())
		cmd.AddCommand(c.newComputePoolCommand())
		cmd.AddCommand(c.newConnectivityTypeCommand())
		cmd.AddCommand(c.newEndpointCommand())
		cmd.AddCommand(c.newRegionCommand())
		cmd.AddCommand(c.newShellCommand(prerunner))
		cmd.AddCommand(c.newStatementCommand())
	}

	return cmd
}

func (c *command) addComputePoolFlag(cmd *cobra.Command) {
	cmd.Flags().String("compute-pool", "", "Flink compute pool ID.")
	pcmd.RegisterFlagCompletionFunc(cmd, "compute-pool", c.autocompleteComputePools)
}

func (c *command) autocompleteComputePools(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	computePools, err := c.V2Client.ListFlinkComputePools(environmentId, "")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(computePools))
	for i, computePool := range computePools {
		suggestions[i] = fmt.Sprintf("%s\t%s", computePool.GetId(), computePool.Spec.GetDisplayName())
	}
	return suggestions
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

func addCmfFlagSet(cmd *cobra.Command) {
	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", `Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.`)
	cmd.Flags().String("client-cert-path", "", `Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.`)
	cmd.Flags().String("certificate-authority-path", "", `Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.`)
}

func (c *command) createContext() context.Context {
	if !c.Config.IsOnPremLogin() {
		return context.Background()
	}
	return context.WithValue(context.Background(), cmfsdk.ContextAccessToken, c.Context.GetAuthToken())
}
