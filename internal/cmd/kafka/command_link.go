package kafka

import (
	"context"
	"io/ioutil"
	"strings"

	linkv1 "github.com/confluentinc/cc-structs/kafka/clusterlink/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	sourceBootstrapServersFlagName     = "source_cluster"
	sourceBootstrapServersPropertyName = "bootstrap.servers"
	configFlagName                     = "config"
	configFileFlagName                 = "config-file"
	dryrunFlagName                     = "dry-run"
	noValidateFlagName                 = "no-validate"
	includeTopicsFlagName              = "include-topics"
)

var (
	keyValueFields      = []string{"Key", "Value"}
	linkFieldsWithTopic = []string{"LinkName", "TopicName"}
	linkFields          = []string{"LinkName"}
)

type keyValueDisplay struct {
	Key   string
	Value string
}

type LinkWriter struct {
	LinkName string
}

type LinkTopicWriter struct {
	LinkName  string
	TopicName string
}

type linkCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner pcmd.PreRunner
}

func NewLinkCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedStateFlagCommand(
		&cobra.Command{
			Use:    "link",
			Hidden: true,
			Short:  "Manages inter-cluster links.",
		},
		prerunner, LinkSubcommandFlags)
	cmd := &linkCommand{
		AuthenticatedStateFlagCommand: cliCmd,
		prerunner:                     prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *linkCommand) init() {
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
	listCmd.Flags().Bool(includeTopicsFlagName, false, "If set, will list mirrored topics for the links returned.")
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
				Code: "ccloud kafka link create my_link --source_cluster myhost:1234\nccloud kafka link create my_link --source_cluster myhost:1234 --config-file ~/myfile.txt",
			},
		),
		RunE: c.create,
		Args: cobra.ExactArgs(1),
	}
	createCmd.Flags().String(sourceBootstrapServersFlagName, "", "Bootstrap-server address for source cluster.")
	createCmd.Flags().String(configFileFlagName, "", "File containing additional comma-separated properties for source cluster.")
	createCmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	createCmd.Flags().Bool(noValidateFlagName, false, "If set, will NOT validate the link to the source cluster before creation.")
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
				Code: "ccloud kafka link update my_link --config \"retention.ms=123456890\"",
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

	includeTopics, err := cmd.Flags().GetBool(includeTopicsFlagName)
	if err != nil {
		return err
	}

	resp, err := c.Client.Kafka.ListLinks(context.Background(), cluster, includeTopics)
	if err != nil {
		return err
	}

	if includeTopics {
		outputWriter, err := output.NewListOutputWriter(
			cmd, linkFieldsWithTopic, linkFieldsWithTopic, linkFieldsWithTopic)
		if err != nil {
			return err
		}

		for _, link := range resp.Links {
			if len(link.Topics) > 0 {
				for topic := range link.Topics {
					outputWriter.AddElement(
						&LinkTopicWriter{LinkName: link.LinkName, TopicName: topic})
				}
			} else {
				outputWriter.AddElement(
					&LinkTopicWriter{LinkName: link.LinkName, TopicName: ""})
			}
		}

		return outputWriter.Out()
	} else {
		outputWriter, err := output.NewListOutputWriter(cmd, linkFields, linkFields, linkFields)
		if err != nil {
			return err
		}

		for _, link := range resp.Links {
			outputWriter.AddElement(&LinkWriter{LinkName: link.LinkName})
		}

		return outputWriter.Out()
	}
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

	skipValidatingLink, err := cmd.Flags().GetBool(noValidateFlagName)
	if err != nil {
		return err
	}

	// Read in extra configs if applicable.
	configFile, err := cmd.Flags().GetString(configFileFlagName)
	if err != nil {
		return err
	}

	var configMap map[string]string
	if configFile != "" {
		configContents, err := ioutil.ReadFile(configFile)
		if err != nil {
			return err
		}

		// Create config map from the argument.
		var linkConfigs []string
		for _, s := range strings.Split(string(configContents), "\n") {
			// Filter out blank lines
			if s != "" {
				linkConfigs = append(linkConfigs, s)
			}
		}
		configMap, err = toMap(linkConfigs)
		if err != nil {
			return err
		}
	} else {
		configMap = make(map[string]string)
	}

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
	createOptions := &linkv1.CreateLinkOptions{ValidateLink: !skipValidatingLink, ValidateOnly: validateOnly}
	err = c.Client.Kafka.CreateLink(context.Background(), cluster, sourceLink, createOptions)

	if err == nil {
		msg := errors.CreatedLinkMsg
		if validateOnly {
			msg = errors.DryRunPrefix + msg
		}
		utils.Printf(cmd, msg, linkName)
	}

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

	if err == nil {
		utils.Printf(cmd, errors.DeletedLinkMsg, link)
	}

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

	if err == nil {
		utils.Printf(cmd, errors.UpdatedLinkMsg, link)
	}

	return err
}
