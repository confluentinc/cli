package schemaregistry

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func (c *command) newCompatibilityCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "compatibility",
		Short:       "Validate schema compatibility.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newCompatibilityValidateCommand(cfg))

	return cmd
}
