package schemaregistry

import (
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/spf13/cobra"
)

func (c *command) newEndpointsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "endpoint",
		Short:       "Manage Schema Registry endpoints.",
		Long:        "Manage the Schema Registry endpoints for the current environment.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	cmd.AddCommand(c.newEndpointsList())

	return cmd
}
