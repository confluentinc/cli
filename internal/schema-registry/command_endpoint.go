package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
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
