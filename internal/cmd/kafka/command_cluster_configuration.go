package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type configurationOut struct {
	Name     string `human:"Name" serialized:"name"`
	Value    string `human:"Value" serialized:"value"`
	ReadOnly bool   `human:"Read-Only" serialized:"read_only"`
}

func (c *clusterCommand) newConfigurationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "configuration",
		Short:       "Manage Kafka cluster configurations.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newConfigurationDescribeCommand())
	cmd.AddCommand(c.newConfigurationListCommand())
	cmd.AddCommand(c.newConfigurationUpdateCommand())

	return cmd
}
