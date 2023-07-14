package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
)

func New(cfg *v1.Config, prerunner pcmd.PreRunner, clientID string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kafka",
		Short: "Manage Apache Kafka.",
	}

	cmd.AddCommand(newAclCommand(cfg, prerunner))
	cmd.AddCommand(newBrokerCommand(prerunner))
	cmd.AddCommand(newClientConfigCommand(prerunner, clientID))
	cmd.AddCommand(newClusterCommand(cfg, prerunner))
	cmd.AddCommand(newConsumerGroupCommand(prerunner))
	cmd.AddCommand(newLinkCommand(cfg, prerunner))
	cmd.AddCommand(newMirrorCommand(prerunner))
	cmd.AddCommand(newPartitionCommand(prerunner))
	cmd.AddCommand(newRegionCommand(prerunner))
	cmd.AddCommand(newReplicaCommand(prerunner))
	cmd.AddCommand(newTopicCommand(cfg, prerunner, clientID))

	dc := dynamicconfig.New(cfg, nil)
	_ = dc.ParseFlagsIntoConfig(cmd)
	if cfg.IsTest || featureflags.Manager.BoolVariation("cli.client_quotas.enable", dc.Context(), v1.CliLaunchDarklyClient, true, false) {
		cmd.AddCommand(newQuotaCommand(prerunner))
	}

	return cmd
}
