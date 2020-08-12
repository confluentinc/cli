package kafka

import (
	"context"
	linkv1 "github.com/confluentinc/cc-structs/kafka/clusterlink/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

const (
	sourceBootstrapServersFlagName     = "source_cluster"
	sourceBootstrapServersPropertyName = "bootstrap.servers"
	configFlagName                     = "config"
	dryrunFlagName                     = "dry-run"
	validateFlagName                   = "validate"
)

var (
	keyValueFields = []string{"Key", "Value"}
)

type keyValueDisplay struct {
	Key   string
	Value string
}

type linkCommand struct {
	*pcmd.AuthenticatedCLICommand
	prerunner pcmd.PreRunner
}

func NewLinkCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:    "link",
			Hidden: true,
			Short:  "Manages inter-cluster links.",
		},
		prerunner)
	cmd := &linkCommand{
		AuthenticatedCLICommand: cliCmd,
		prerunner:               prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *linkCommand) init() {
	c.Command.PersistentFlags().String("cluster", "", "Kafka cluster ID.")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List previously created cluster links.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List every link",
				Code: "ccloud kafka link list",
			},
		),
		RunE: c.list,
		Args: cobra.NoArgs,
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	// Note: this is subject to change as we iterate on options for how to specify a source cluster.
	createCmd := &cobra.Command{
		Use:   "create <link-name>",
		Short: "Create a new cluster link.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a cluster link, using supplied source URL and properties.",
				Code: "ccloud kafka link create my_link --source_cluster myhost:1234\nccloud kafka link create my_link --source_cluster myhost:1234 --config \"key1=val1,key2=val2\"",
			},
		),
		RunE: c.create,
		Args: cobra.ExactArgs(1),
	}
	createCmd.Flags().String(sourceBootstrapServersFlagName, "", "Bootstrap-server address for source cluster.")
	createCmd.Flags().StringSlice(configFlagName, nil, "Additional comma-separated properties for source cluster.")
	createCmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	createCmd.Flags().Bool(validateFlagName, false, "If set, will validate the link to the source cluster before creation.")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <link-name>",
		Short: "Delete a previously created cluster link.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Deletes a cluster link.",
				Code: "ccloud kafka link delete my_link",
			},
		),
		RunE: c.delete,
		Args: cobra.ExactArgs(1),
	}
	c.AddCommand(deleteCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <link-name>",
		Short: "Describes a previously created cluster link.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describes a cluster link.",
				Code: "ccloud kafka link describe my_link",
			},
		),
		RunE: c.describe,
		Args: cobra.ExactArgs(1),
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	c.AddCommand(describeCmd)

	// Note: this can change as we decide how to present this modification interface (allowing multiple properties, allowing override and delete, etc).
	updateCmd := &cobra.Command{
		Use:   "update <link-name>",
		Short: "Updates a property for a previously created cluster link.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Updates a property for a cluster link.",
				Code: "ccloud kafka link update my_link --key retention.ms --value 123456890",
			},
		),
		RunE: c.update,
		Args: cobra.ExactArgs(1),
	}
	updateCmd.Flags().StringSlice("config", nil, "A comma-separated list of topics. Configuration ('key=value') overrides for the topic being created.")
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)
}

func (c *linkCommand) list(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}
	resp, err := c.Client.Kafka.ListLinks(context.Background(), cluster)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, []string{"LinkName"}, []string{"LinkName"}, []string{"LinkName"})
	if err != nil {
		return err
	}
	type LinkWriter struct {
		LinkName string
	}
	for _, link := range resp {
		outputWriter.AddElement(&LinkWriter{LinkName: link})
	}
	return outputWriter.Out()
}

func (c *linkCommand) create(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}

	linkName := args[0]

	bootstrapServers, err := cmd.Flags().GetString(sourceBootstrapServersFlagName)
	if err != nil {
		return err
	}

	validateOnly, err := cmd.Flags().GetBool(dryrunFlagName)
	if err != nil {
		return err
	}

	validateLink, err := cmd.Flags().GetBool(validateFlagName)
	if err != nil {
		return err
	}

	linkConfigs, err := cmd.Flags().GetStringSlice(configFlagName)
	if err != nil {
		return err
	}
	// Create config map from the argument.
	configMap, err := toMap(linkConfigs)

	// The `source` argument is a convenience; we package everything into properties for the source cluster.
	configMap[sourceBootstrapServersPropertyName] = bootstrapServers

	if err != nil {
		return err
	}
	sourceLink := &linkv1.ClusterLink{
		LinkName:  linkName,
		ClusterId: "",
		Configs:   configMap,
	}
	createOptions := &linkv1.CreateLinkOptions{ValidateLink: validateLink, ValidateOnly: validateOnly}
	err = c.Client.Kafka.CreateLink(context.Background(), cluster, sourceLink, createOptions)
	return err
}

func (c *linkCommand) delete(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}

	link := args[0]
	deletionOptions := &linkv1.DeleteLinkOptions{}
	err = c.Client.Kafka.DeleteLink(context.Background(), cluster, link, deletionOptions)
	return err
}

func (c *linkCommand) describe(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}

	link := args[0]
	resp, err := c.Client.Kafka.DescribeLink(context.Background(), cluster, link)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, keyValueFields, keyValueFields, keyValueFields)
	if err != nil {
		return err
	}

	for k, v := range resp.Properties {
		outputWriter.AddElement(&keyValueDisplay{
			Key:   k,
			Value: v,
		})
	}
	return outputWriter.Out()
}

func (c *linkCommand) update(cmd *cobra.Command, args []string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}

	link := args[0]
	configs, err := cmd.Flags().GetStringSlice(configFlagName)
	if err != nil {
		return err
	}
	configMap, err := toMap(configs)
	if err != nil {
		return err
	}

	config := &linkv1.LinkProperties{
		Properties: configMap,
	}
	alterOptions := &linkv1.AlterLinkOptions{}
	err = c.Client.Kafka.AlterLink(context.Background(), cluster, link, config, alterOptions)

	return err
}
