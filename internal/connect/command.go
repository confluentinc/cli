package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

type connectOut struct {
	Id     string `human:"ID" serialized:"id"`
	Name   string `human:"Name" serialized:"name"`
	Status string `human:"Status" serialized:"status"`
	Type   string `human:"Type" serialized:"type"`
	Trace  string `human:"Trace,omitempty" serialized:"trace,omitempty"`
}

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "connect",
		Short:       "Manage Kafka Connect.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(newClusterCommand(cfg, prerunner))
	cmd.AddCommand(newCustomPluginCommand(prerunner))
	cmd.AddCommand(newEventCommand(prerunner))
	cmd.AddCommand(newPluginCommand(cfg, prerunner))

	return cmd
}
