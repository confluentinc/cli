package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

type endpointOut struct {
	IsCurrent              bool   `human:"Current" serialized:"is_current"`
	Endpoint               string `human:"Endpoint" serialized:"endpoint"`
	KafkaBootstrapEndpoint string `human:"Kafka Bootstrap Endpoint" serialized:"kafka_bootstrap_endpoint"`
	HttpEndpoint           string `human:"Http Endpoint" serialized:"http_endpoint"`
	ConnectionType         string `human:"Connection Type" serialized:"connection_type"`
}

func (c *clusterCommand) newEndpointCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "endpoint",
		Short: "Manage Kafka cluster endpoint.",
	}

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newEndpointListCommand())
		cmd.AddCommand(c.newEndpointUseCommand())
	}

	return cmd
}
