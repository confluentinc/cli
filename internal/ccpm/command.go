package ccpm

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "ccpm",
		Short:       "Manage Custom Connect Plugin Management (CCPM).",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(newPluginCommand(prerunner))
	cmd.AddCommand(newPresignedUrlCommand(prerunner))
	cmd.AddCommand(newVersionCommand(prerunner))

	return cmd
}
