package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

type endpointOut struct {
	IsCurrent bool   `human:"Current" serialized:"is_current"`
	Endpoint  string `human:"Endpoint" serialized:"endpoint"`
	Cloud     string `human:"Cloud" serialized:"cloud"`
	Region    string `human:"Region" serialized:"region"`
	Type      string `human:"Type" serialized:"type"`
}

func (c *clusterCommand) newEndpointCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "endpoint",
		Short: "Manage Kafka cluster endpoint.",
	}

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newEndpointListCommand())
		cmd.AddCommand(c.newEndpointUseCommand())
	} else {
		c.PersistentPreRunE = prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand)

		// QUESTION: do we even support endpoint on-prem? Assuming so here...
		cmd.AddCommand(c.newEndpointListCommandOnPrem())
		cmd.AddCommand(c.newEndpointUseCommandOnPrem())
	}

	return cmd
}
