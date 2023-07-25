package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func (c *command) newConfigCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "config",
		Short:       "Manage Schema Registry configuration.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newConfigDescribeCommand(cfg))

	return cmd
}
