package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
}

type connectOut struct {
	Id     string `human:"ID" serialized:"id"`
	Name   string `human:"Name" serialized:"name"`
	Status string `human:"Status" serialized:"status"`
	Type   string `human:"Type" serialized:"type"`
	Trace  string `human:"Trace,omitempty" serialized:"trace,omitempty"`
}

func New(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "connect",
		Short:       "Manage Kafka Connect.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &command{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	c.AddCommand(newClusterCommand(cfg, prerunner))
	c.AddCommand(newEventCommand(prerunner))
	c.AddCommand(newPluginCommand(prerunner))

	return c.Command
}
