package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "flink",
		Short:       "Manage Apache Flink.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newComputePoolCommand(cfg))
	cmd.AddCommand(c.newRegionCommand())
	cmd.AddCommand(c.newShellCommand(cfg, prerunner))
	cmd.AddCommand(c.newStatementCommand())

	dc := dynamicconfig.New(cfg, nil)
	_ = dc.ParseFlagsIntoConfig(cmd)
	if cfg.IsTest || featureflags.Manager.BoolVariation("cli.flink.open_preview", dc.Context(), v1.CliLaunchDarklyClient, true, false) {
		cmd.AddCommand(c.newIamBindingCommand())
	}

	return cmd
}

func (c *command) addDatabaseFlag(cmd *cobra.Command) {
	cmd.Flags().String("database", "", "The database against which the statement will run. For example, the display name of a Kafka cluster.")

	pcmd.RegisterFlagCompletionFunc(cmd, "database", func(cmd *cobra.Command, args []string) []string {
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
			suggestions[i] = cluster.Spec.GetDisplayName()
		}

		return suggestions
	})
}
