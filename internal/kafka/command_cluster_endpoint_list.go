package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newEndpointListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		RunE:  c.endpointList,
		Short: "List Kafka Cluster endpoint.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List the available Kafka Cluster endpoints with current cloud provider and region.",
				Code: "confluent kafka cluster endpoint list",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) endpointList(cmd *cobra.Command, args []string) error {
	// Add logic for displaying endpoint, layer by later
	// check current environment... cloud... region...
	// PNI -> privatelink -> public
	return nil
}
