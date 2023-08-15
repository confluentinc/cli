package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *command) newConfigCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "config",
		Short:       "Manage Schema Registry configuration.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newConfigDescribeCommand(cfg))

	return cmd
}
