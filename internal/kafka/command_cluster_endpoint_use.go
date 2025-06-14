package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newEndpointUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use",
		Short: "Use a Kafka Cluster endpoint.",
		Long:  "Use a Kafka Cluster endpoint as active endpoint for all subsequent Kafka Cluster commands in current environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.endpointUse,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Use "https://lkc-s1232.us-west-2.aws.private.confluent.cloud:443" for subsequent Kafka Cluster commands.`,
				Code: `confluent kafka cluster endpoint use "https://lkc-s1232.us-west-2.aws.private.confluent.cloud:443"`,
			},
		),
	}

	return cmd
}

func (c *command) endpointUse(cmd *cobra.Command, args []string) error {
	// TODO: add logic and endpoint validation here! Sanity checks!

	return nil
}
