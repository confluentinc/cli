package connect

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
}

type connectCreateDisplay struct {
	ConnectorName string `json:"name" yaml:"name"`
	Id            string `json:"id" yaml:"id"`
	Trace         string `json:"error_trace,omitempty" yaml:"error_trace,omitempty"`
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

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "connect",
		Short:       "Manage Kafka Connect.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &command{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	c.AddCommand(newClusterCommand(prerunner))
	c.AddCommand(c.newCreateCommand())
	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(newEventCommand(prerunner))
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newPauseCommand())
	c.AddCommand(newPluginCommand(prerunner))
	c.AddCommand(c.newResumeCommand())
	c.AddCommand(c.newUpdateCommand())

	return c.Command
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteConnectors()
}

func (c *command) autocompleteConnectors() []string {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return nil
	}

	connectors, _, err := c.V2Client.ListConnectorsWithExpansions(c.EnvironmentId(), kafkaCluster.ID, "status,info,id")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(connectors))
	i := 0
	for _, connector := range connectors {
		suggestions[i] = fmt.Sprintf("%s\t%s", connector.Id.GetId(), connector.Info.GetName())
		i++
	}
	return suggestions
}
