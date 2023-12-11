package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *command) newDekVersionCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "version",
		Short:       "Manage Schema Registry DEK versions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newDekVersionListCommand(cfg))

	return cmd
}
