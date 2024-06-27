package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *command) newSchemaReferenceCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "reference",
		Short:       "Manage Schema Registry schema references.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newSchemaReferenceListCommand(cfg))

	return cmd
}
