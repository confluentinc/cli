package connect

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
	completableChildren     []*cobra.Command
	completableFlagChildren map[string][]*cobra.Command
	analyticsClient         analytics.Client
}

type connectorDescribeDisplay struct {
	Name   string `json:"name" yaml:"name"`
	ID     string `json:"id" yaml:"id"`
	Status string `json:"status" yaml:"status"`
	Type   string `json:"type" yaml:"type"`
	Trace  string `json:"trace,omitempty" yaml:"trace,omitempty"`
}

var (
	listFields           = []string{"ID", "Name", "Status", "Type", "Trace"}
	listStructuredLabels = []string{"id", "name", "status", "type", "trace"}
)

// New returns the default command object for interacting with Connect.
func New(prerunner pcmd.PreRunner, analyticsClient analytics.Client) *command {
	cmd := &cobra.Command{
		Use:         "connect",
		Short:       "Manage Kafka Connect.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &command{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner, subcommandFlags),
		analyticsClient:               analyticsClient,
	}

	createCmd := c.newCreateCommand()
	describeCmd := c.newDescribeCommand()
	deleteCmd := c.newDeleteCommand()
	listCmd := c.newListCommand()
	pauseCmd := c.newPauseCommand()
	resumeCmd := c.newResumeCommand()
	updateCmd := c.newUpdateCommand()

	c.AddCommand(newClusterCommand(prerunner))
	c.AddCommand(createCmd)
	c.AddCommand(deleteCmd)
	c.AddCommand(describeCmd)
	c.AddCommand(newEventCommand(prerunner))
	c.AddCommand(listCmd)
	c.AddCommand(pauseCmd)
	c.AddCommand(newPluginCommand(prerunner))
	c.AddCommand(resumeCmd)
	c.AddCommand(updateCmd)

	c.completableChildren = []*cobra.Command{deleteCmd, describeCmd, pauseCmd, resumeCmd, updateCmd}
	c.completableFlagChildren = map[string][]*cobra.Command{"cluster": {createCmd}}

	return c
}
