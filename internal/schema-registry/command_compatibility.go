package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

func (c *command) newCompatibilityCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "compatibility",
		Short:       "Validate schema compatibility.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newCompatibilityValidateCommand(cfg))

	return cmd
}
