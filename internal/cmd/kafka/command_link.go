package kafka

import (
	"context"
	"fmt"
	"github.com/antihax/optional"
	linkv1 "github.com/confluentinc/cc-structs/kafka/clusterlink/v1"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	sourceBootstrapServersFlagName     = "source-cluster"
	sourceClusterIdName                = "source-cluster-id"
	sourceBootstrapServersPropertyName = "bootstrap.servers"
	configFileFlagName                  = "config-file"
	dryrunFlagName                     = "dry-run"
	noValidateFlagName                 = "no-validate"
	includeTopicsFlagName              = "include-topics"
	linkFlagName                       = "link-name"
)

var (
	keyValueFields      = []string{"Key", "Value"}
	linkFieldsWithTopic = []string{"LinkName", "TopicName"}
	linkFields          = []string{"LinkName"}
	linkConfigFields     = []string{"ConfigName", "ConfigValue", "ReadOnly", "Sensitive", "Source", "Synonyms"}
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

type LinkConfigWriter struct {
	ConfigName  string
	ConfigValue string
	ReadOnly    bool
	Sensitive   bool
	Source      string
	Synonyms    []string
}

type linkCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner pcmd.PreRunner
}

func NewLinkCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedStateFlagCommand(
		&cobra.Command{
			Use:    "link",
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
				Text: "List every link.",
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
	createCmd.Flags().String(sourceBootstrapServersFlagName, "", "Bootstrap-server address of the source cluster.")
	createCmd.Flags().String(sourceClusterIdName, "", "Source cluster ID.")
	createCmd.Flags().String(configFileFlagName, "", "Name of the file containing link config overrides. " +
		"Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")
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
		Short: "Update link configs.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Updates configs for the cluster link.",
				Code: "ccloud kafka link update my_link --config \"retention.ms=123456890\"",
			},
		),
		RunE: c.update,
		Args: cobra.ExactArgs(1),
	}
	updateCmd.Flags().String(configFileFlagName, "", "Name of the file containing link config overrides. " +
		"Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	listConfigCmd := &cobra.Command{
		Use:   "list-configs <link-name>",
		Short: "List all configs of the link.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all configs of the link.",
				Code: "ccloud kafka link list-configs",
			},
		),
		RunE: c.listConfigs,
		Args: cobra.ExactArgs(1),
	}
	listConfigCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listConfigCmd.Flags().SortFlags = false
	c.AddCommand(listConfigCmd)
}

func (c *linkCommand) listConfigs(cmd *cobra.Command, args []string) error {
	linkName := args[0]
	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST == nil {
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	listLinkConfigsResponseDataList, _, err := kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksLinkNameConfigsGet(
		kafkaREST.Context, lkc, linkName)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, linkConfigFields, linkConfigFields, linkConfigFields)
	if err != nil {
		return err
	}

	for _, config := range listLinkConfigsResponseDataList.Data {
		outputWriter.AddElement(&LinkConfigWriter{
			ConfigName:  config.Name,
			ConfigValue: config.Value,
			ReadOnly:    config.ReadOnly,
			Sensitive:   config.Sensitive,
			Source:      config.Source,
			Synonyms:    config.Synonyms,
		})
	}

	return outputWriter.Out()
}

func (c *linkCommand) list(cmd *cobra.Command, args []string) error {
	includeTopics, err := cmd.Flags().GetBool(includeTopicsFlagName)
	if err != nil {
		return err
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST == nil {
		// Fall back to use kafka-api if the cluster doesn't support rest proxy
		return c.listWithKafkaApi(cmd, includeTopics)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	listLinksRespDataList, httpResp, err := kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksGet(
		kafkaREST.Context, lkc)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	if includeTopics {
		outputWriter, err := output.NewListOutputWriter(
			cmd, linkFieldsWithTopic, linkFieldsWithTopic, linkFieldsWithTopic)
		if err != nil {
			return err
		}

		for _, link := range listLinksRespDataList.Data {
			if len(link.TopicsNames) > 0 {
				for _, topic := range link.TopicsNames {
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

		for _, link := range listLinksRespDataList.Data {
			outputWriter.AddElement(&LinkWriter{LinkName: link.LinkName})
		}

		return outputWriter.Out()
	}
}

// Will be deprecated soon
func (c *linkCommand) listWithKafkaApi(cmd *cobra.Command, includeTopics bool) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
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
	linkName := args[0]

	bootstrapServers, err := cmd.Flags().GetString(sourceBootstrapServersFlagName)
	if err != nil {
		return err
	}

	sourceClusterId, err := cmd.Flags().GetString(sourceClusterIdName)
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

	configMap, err := readConfigsFromFile(configFile)
	if err != nil {
		return err
	}

	// The `source` argument is a convenience; we package everything into properties for the Source cluster.
	configMap[sourceBootstrapServersPropertyName] = bootstrapServers

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST == nil {
		// Fall back to use kafka-api if the cluster doesn't support rest proxy
		return c.createWithKafkaApi(cmd, linkName, configMap, skipValidatingLink, validateOnly)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	createLinkOpt := &kafkarestv3.ClustersClusterIdLinksPostOpts{
		ValidateOnly: optional.NewBool(validateOnly),
		ValidateLink: optional.NewBool(!skipValidatingLink),
		CreateLinkRequestData: optional.NewInterface(&kafkarestv3.CreateLinkRequestData{
			SourceClusterId: sourceClusterId,
			Configs: toCreateTopicConfigs(configMap),
		}),
	}

	_, err = kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksPost(
		kafkaREST.Context, lkc, linkName, createLinkOpt)

	if err == nil {
		msg := errors.CreatedLinkMsg
		if validateOnly {
			msg = errors.DryRunPrefix + msg
		}
		utils.Printf(cmd, msg, linkName)
	}

	return err
}

// Will be deprecated soon
func (c* linkCommand) createWithKafkaApi(
	cmd *cobra.Command, linkName string, configMap map[string]string, skipValidatingLink bool, validateOnly bool) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
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
	linkName := args[0]
	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST == nil {
		// Fall back to use kafka-api if the cluster doesn't support rest proxy
		return c.deleteWithKafkaApi(cmd, linkName)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	_, err = kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksLinkNameDelete(kafkaREST.Context, lkc, linkName)
	if err == nil {
		utils.Printf(cmd, errors.DeletedLinkMsg, linkName)
	}

	return err
}

// Will be deprecated soon
func (c *linkCommand) deleteWithKafkaApi(cmd *cobra.Command, linkName string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}

	deletionOptions := &linkv1.DeleteLinkOptions{}
	err = c.Client.Kafka.DeleteLink(context.Background(), cluster, linkName, deletionOptions)
	if err == nil {
		utils.Printf(cmd, errors.DeletedLinkMsg, linkName)
	}

	return err
}

func (c *linkCommand) describe(cmd *cobra.Command, args []string) error {
	linkName := args[0]
	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST == nil {
		fmt.Println("rest proxy not available")
		// Fall back to use kafka-api if the cluster doesn't support rest proxy
		return c.describeWithKafkaApi(cmd, linkName)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	listLinksResponseData, _, err := kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksLinkNameGet(
		kafkaREST.Context, lkc, linkName)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, keyValueFields, keyValueFields, keyValueFields)
	if err != nil {
		return err
	}

	outputWriter.AddElement(&keyValueDisplay{
		Key: "ClusterId",
		Value: listLinksResponseData.SourceClusterId,
	})

	outputWriter.AddElement(&keyValueDisplay{
		Key: "LinkName",
		Value: listLinksResponseData.LinkName,
	})

	outputWriter.AddElement(&keyValueDisplay{
		Key: "LinkId",
		Value: listLinksResponseData.LinkId,
	})


	listLinkConfigsRespData, _, err := kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksLinkNameConfigsGet(
		kafkaREST.Context, lkc, linkName)
	if err != nil {
		return err
	}

	for _, config := range listLinkConfigsRespData.Data {
		outputWriter.AddElement(
			&keyValueDisplay{
				Key:   config.Name,
				Value: config.Value,
			})
	}

	return outputWriter.Out()
}

// Will be deprecated soon
func (c *linkCommand) describeWithKafkaApi(cmd *cobra.Command, link string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}

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
	linkName := args[0]
	configFile, err := cmd.Flags().GetString(configFileFlagName)
	if err != nil {
		return err
	}

	configsMap, err := readConfigsFromFile(configFile)
	if err != nil {
		return err
	}

	if len(configsMap) == 0 {
		return errors.New(errors.EmptyConfigErrorMsg)
	}

	kafkaREST, _ := c.GetKafkaREST()
	if kafkaREST == nil {
		// Fall back to use kafka-api if the cluster doesn't support rest proxy
		return c.updateWithKafkaApi(cmd, linkName, configsMap)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	kafkaRestConfigs := toAlterConfigBatchRequestData(configsMap)

	_, err = kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksLinkNameConfigsalterPut(
		kafkaREST.Context, lkc, linkName,
		&kafkarestv3.ClustersClusterIdLinksLinkNameConfigsalterPutOpts{
			AlterConfigBatchRequestData: optional.NewInterface(
				kafkarestv3.AlterConfigBatchRequestData{Data: kafkaRestConfigs}),
		})
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.UpdatedLinkMsg, linkName)
	return nil
}

// Will be deprecated soon
func (c *linkCommand) updateWithKafkaApi(cmd *cobra.Command, linkName string, configMap map[string]string) error {
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}

	config := &linkv1.LinkProperties{
		Properties: configMap,
	}
	alterOptions := &linkv1.AlterLinkOptions{}
	err = c.Client.Kafka.AlterLink(context.Background(), cluster, linkName, config, alterOptions)

	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.UpdatedLinkMsg, linkName)
	return nil
}
