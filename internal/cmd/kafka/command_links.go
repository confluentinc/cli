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
	sourceBootstrapServersFlagName = "source"
)

var (
	keyValueFields = []string{"Key", "Value"}
)

type keyValueDisplay struct {
	Key string
	Value string
}

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
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	// Note: this is subject to change as we iterate on options for how to specify a source cluster.
	createCmd := &cobra.Command{
		Use: "create <name>",
		Short: "Create a new cluster-link.",
		Example: `
Create a cluster-link.

::

        ccloud kafka links create MyLink --source myhost:1234`,
		RunE: c.create,
		Args: cobra.ExactArgs(1),
	}
	createCmd.Flags().String(sourceBootstrapServersFlagName, "", "Bootstrap-servers address for source cluster.")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use: "delete <name>",
		Short: "Delete a previously created cluster-link.",
		Example: `
Deletes a cluster-link.

::

        ccloud kafka links delete Mylink`,
		RunE: c.delete,
		Args: cobra.ExactArgs(1),
	}
	c.AddCommand(deleteCmd)

	describeCmd := &cobra.Command{
		Use: "describe <name>",
		Short: "Describes a previously created cluster-link.",
		Example: `
Describes a cluster-link.

::

        ccloud kafka links describe MyLink`,
        RunE: c.describe,
        Args: cobra.ExactArgs(1),
	}
	c.AddCommand(describeCmd)

	// Note: this can change as we decide how to present this modification interface (allowing multiple properties, allowing override and delete, etc).
	alterCmd := &cobra.Command{
		Use: "alter <name> <k> <v>",
		Short: "Alters a particular property for a previously created cluster-link.",
		Example: `
Alters a property for a cluster-link.

::

        ccloud kafka links alter MyLink retention.ms 123456890`,
		RunE: c.alter,
		Args: cobra.ExactArgs(3),
	}
	c.AddCommand(alterCmd)
}

func (c *linksCommand) list(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	resp, err := c.Client.Kafka.ListLinks(context.Background(), cluster)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, []string{"Name"}, []string{"Name"}, []string{"Name"})
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	for _, link := range resp {
		outputWriter.AddElement(link)
	}
	return outputWriter.Out()
}

func (c *linksCommand) create(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	link := &kafkav1.Link{
		Name:                 args[0],
	}
	// Handle multiple options for source-cluster specification here.
	bootstrapServers, err := cmd.Flags().GetString(sourceBootstrapServersFlagName)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	sourceCluster := &kafkav1.LinkSourceCluster{
		BootstrapServers:     bootstrapServers,
		Configs:              nil,
	}
	err = c.Client.Kafka.CreateLink(context.Background(), cluster, link, sourceCluster)
	return errors.HandleCommon(err, cmd)
}

func (c *linksCommand) delete(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	link := &kafkav1.Link{
		Name:                 args[0],
	}
	err = c.Client.Kafka.DeleteLink(context.Background(), cluster, link)
	return errors.HandleCommon(err, cmd)
}

func (c *linksCommand) describe(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	link := &kafkav1.Link{
		Name: args[0],
	}
	resp, err := c.Client.Kafka.DescribeLink(context.Background(), cluster, link)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, keyValueFields, keyValueFields, keyValueFields)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	for k, v := range resp.Properties {
		outputWriter.AddElement(&keyValueDisplay{
			Key: k,
			Value: v,
		})
	}
	return outputWriter.Out()
}

func (c *linksCommand) alter(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	link := &kafkav1.Link{
		Name: args[0],
	}
	key := args[1]
	value := args[2]
	config := &kafkav1.LinkDescription{
		Properties: map[string]string{key: value},
	}
	err = c.Client.Kafka.AlterLink(context.Background(), cluster, link, config)

	return errors.HandleCommon(err, cmd)
}
