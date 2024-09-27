package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

var fields = []string{"private", "public"}

func (c *command) newConnectivityTypeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "connectivity-type",
		Short:       "Manage Flink connectivity type.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newUseCommand())

	return cmd
}
