package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
)

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kafka",
		Short: "Manage Apache Kafka.",
	}

	cmd.AddCommand(newAclCommand(cfg, prerunner))
	cmd.AddCommand(newBrokerCommand(prerunner))
	cmd.AddCommand(newClientConfigCommand(cfg, prerunner))
	cmd.AddCommand(newClusterCommand(cfg, prerunner))
	cmd.AddCommand(newConsumerCommand(cfg, prerunner))
	cmd.AddCommand(newConsumerGroupCommand(cfg, prerunner))
	cmd.AddCommand(newLinkCommand(cfg, prerunner))
	cmd.AddCommand(newMirrorCommand(prerunner))
	cmd.AddCommand(newPartitionCommand(prerunner))
	cmd.AddCommand(newRegionCommand(prerunner))
	cmd.AddCommand(newReplicaCommand(prerunner))
	cmd.AddCommand(newTopicCommand(cfg, prerunner))

	dc := dynamicconfig.New(cfg, nil)
	_ = dc.ParseFlagsIntoConfig(cmd)
	if cfg.IsTest || featureflags.Manager.BoolVariation("cli.client_quotas.enable", dc.Context(), config.CliLaunchDarklyClient, true, false) {
		cmd.AddCommand(newQuotaCommand(prerunner))
	}

	return cmd
}
