package kafka

import (
	"context"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

const (
	sourceBootstrapServersFlagName = "source-bootstrap-servers"
)

type linksCommand struct {
	*pcmd.AuthenticatedCLICommand
	prerunner pcmd.PreRunner
}

func NewLinksCommand(prerunner pcmd.PreRunner, config *v3.Config) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use: "links",
			Short: "Manage inter-cluster links.",
		},
		config, prerunner)
	cmd := &linksCommand{
		AuthenticatedCLICommand: cliCmd,
		prerunner:               prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *linksCommand) init() {
	listCmd := &cobra.Command{
		Use: "list",
		Short: "List previously created cluster-links.",
		Example: `
List all links.

::

        ccloud kafka links list`,
		RunE: c.list,
		Args: cobra.NoArgs}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	c.AddCommand(listCmd)

	// Note: this is subject to change as we iterate on options for how to specify a source cluster.
	createCmd := &cobra.Command{
		Use: "create <link-name> <source-cluster>",
		Short: "Create a new cluster-link.",
		Example: `
Create a cluster-link.

::

        ccloud kafka links create MyLink --source-bootstrap-servers myhost:1234`,
		RunE: c.create,
		Args: cobra.ExactArgs(1),
	}
	createCmd.Flags().String(sourceBootstrapServersFlagName, "", "Bootstrap-servers address for source cluster")
	c.AddCommand(createCmd)
}

func (c *linksCommand) list(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		errors.HandleCommon(err, cmd)
	}
	resp, err := c.Client.Kafka.ListLinks(context.Background(), cluster)
	if err != nil {
		errors.HandleCommon(err, cmd)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, []string{"Name"}, []string{"Name"}, []string{"Name"})
	if err != nil {
		errors.HandleCommon(err, cmd)
	}
	for _, link := range resp {
		outputWriter.AddElement(link)
	}
	return outputWriter.Out()
}

func (c *linksCommand) create(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		errors.HandleCommon(err, cmd)
	}

	link := &kafkav1.Link{
		Name:                 args[0],
	}
	// Handle multiple options for source-cluster specification here.
	bootstrapServers, err := cmd.Flags().GetString(sourceBootstrapServersFlagName)
	if err != nil {
		errors.HandleCommon(err, cmd)
	}
	sourceCluster := &kafkav1.LinkSourceCluster{
		BootstrapServers:     bootstrapServers,
		Configs:              nil,
	}
	err = c.Client.Kafka.CreateLink(context.Background(), cluster, link, sourceCluster)

	return errors.HandleCommon(err, cmd)
}
