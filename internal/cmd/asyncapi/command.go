package asyncapi

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "asyncapi",
		Short:       "Manage AsyncAPI document tooling.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	cmd.AddCommand(newExportCommand(prerunner))

	return cmd
}
