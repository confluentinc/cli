package kafka

import (
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	pkafka "github.com/confluentinc/cli/internal/pkg/kafka"
	"github.com/confluentinc/cli/internal/pkg/log"
)

const (
	singleZone = "single-zone"
	multiZone  = "multi-zone"
)

type clusterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	logger              *log.Logger
	completableChildren []*cobra.Command
	analyticsClient     analytics.Client
}

func newClusterCommand(cfg *v1.Config, prerunner pcmd.PreRunner, analyticsClient analytics.Client) *clusterCommand {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Manage Kafka clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &clusterCommand{analyticsClient: analyticsClient}

	if cfg.IsCloudLogin() {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)
	} else {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner)
	}

	deleteCmd := c.newDeleteCommand(cfg)
	describeCmd := c.newDescribeCommand(cfg)
	updateCmd := c.newUpdateCommand(cfg)
	useCmd := c.newUseCommand(cfg)

	c.AddCommand(c.newCreateCommand(cfg))
	c.AddCommand(deleteCmd)
	c.AddCommand(describeCmd)
	c.AddCommand(updateCmd)
	c.AddCommand(useCmd)

	if cfg.IsCloudLogin() {
		c.AddCommand(c.newListCommand())
	} else {
		c.AddCommand(c.newListCommandOnPrem())
	}

	c.completableChildren = []*cobra.Command{deleteCmd, describeCmd, updateCmd, useCmd}

	return c
}

func (c *clusterCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteClusters(c.EnvironmentId(), c.Client)
}

func (c *clusterCommand) Cmd() *cobra.Command {
	return c.Command
}

func (c *clusterCommand) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	clusters, err := pkafka.ListKafkaClusters(c.Client, c.EnvironmentId())
	if err != nil {
		return suggestions
	}
	for _, cluster := range clusters {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        cluster.Id,
			Description: cluster.Name,
		})
	}
	return suggestions
}

func (c *clusterCommand) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}
