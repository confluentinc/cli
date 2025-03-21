package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newRegionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "region",
		Short:       "Manage Flink regions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newRegionListCommand())
	cmd.AddCommand(c.newRegionUseCommand())

	return cmd
}
