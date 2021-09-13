package kafka

// TODO: wrap all link / mirror commands with kafka rest error
import (
	"fmt"

	"github.com/antihax/optional"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	apiKeyFlagName                     = "source-api-key"
	apiSecretFlagName                  = "source-api-secret"
	sourceBootstrapServersFlagName     = "source-bootstrap-server"
	sourceClusterIdFlagName            = "source-cluster-id"
	sourceBootstrapServersPropertyName = "bootstrap.servers"
	securityProtocalPropertyName       = "security.protocol"
	saslMechanismPropertyName          = "sasl.mechanism"
	saslJaasConfigPropertyName         = "sasl.jaas.config"
	configFileFlagName                 = "config-file"
	dryrunFlagName                     = "dry-run"
	noValidateFlagName                 = "no-validate"
	includeTopicsFlagName              = "include-topics"
	linkFlagName                       = "link"
)

var (
	listLinkFieldsIncludeTopics           = []string{"LinkName", "TopicName", "SourceClusterId"}
	structuredListLinkFieldsIncludeTopics = camelToSnake(listLinkFieldsIncludeTopics)
	humanListLinkFieldsIncludeTopics      = camelToSpaced(listLinkFieldsIncludeTopics)
	listLinkFields                        = []string{"LinkName", "SourceClusterId"}
	structuredListLinkFields              = camelToSnake(listLinkFields)
	humanListLinkFields                   = camelToSpaced(listLinkFields)
	describeLinkConfigFields              = []string{"ConfigName", "ConfigValue", "ReadOnly", "Sensitive", "Source", "Synonyms"}
	structuredDescribeLinkConfigFields    = camelToSnake(describeLinkConfigFields)
	humanDescribeLinkConfigFields         = camelToSpaced(describeLinkConfigFields)
)

type LinkTopicWriter struct {
	LinkName        string
	TopicName       string
	SourceClusterId string
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
			Use:         "link",
			Short:       "Manages inter-cluster links.",
			Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
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
		Use:   "create <link>",
		Short: "Create a new cluster link.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a cluster link, using supplied source URL and properties.",
				Code: "ccloud kafka link create my_link --source-cluster-id lkc-abced " +
					"--source-bootstrap-server myhost:1234 --config-file ~/myfile.txt \n" +
					"ccloud kafka link create my_link --source-cluster-id lkc-abced " +
					"--source-bootstrap-server myhost:1234 --source-api-key abcde --source-api-secret 88888 \n",
			},
		),
		RunE: c.create,
		Args: cobra.ExactArgs(1),
	}
	createCmd.Flags().String(sourceBootstrapServersFlagName, "", "Bootstrap-server address of the source cluster.")
	createCmd.Flags().String(sourceClusterIdFlagName, "", "Source cluster ID.")
	check(createCmd.MarkFlagRequired(sourceBootstrapServersFlagName))
	check(createCmd.MarkFlagRequired(sourceClusterIdFlagName))
	createCmd.Flags().String(apiKeyFlagName, "", "An API key for the source cluster. "+
		"If specified, the destination cluster will use SASL_SSL/PLAIN as its mechanism for the source cluster authentication. "+
		"If you wish to use another authentication mechanism, please do NOT specify this flag, "+
		"and add the security configs in the config file. "+
		"Must be used with --source-api-secret.")
	createCmd.Flags().String(apiSecretFlagName, "", "An API secret for the source cluster. "+
		"If specified, the destination cluster will use SASL_SSL/PLAIN as its mechanism for the source cluster authentication. "+
		"If you wish to use another authentication mechanism, please do NOT specify this flag, "+
		"and add the security configs in the config file. "+
		"Must be used with --source-api-key.")
	createCmd.Flags().String(configFileFlagName, "", "Name of the file containing link config overrides. "+
		"Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")
	createCmd.Flags().Bool(dryrunFlagName, false, "If set, will NOT actually create the link, but simply validates it.")
	createCmd.Flags().Bool(noValidateFlagName, false, "If set, will NOT validate the link to the source cluster before creation.")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <link>",
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
		Use:   "describe <link>",
		Short: "Describe a previously created cluster link.",
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
		Use:   "update <link>",
		Short: "Update link configs.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Updates configs for the cluster link.",
				Code: "ccloud kafka link update my_link --config-file ~/config.txt",
			},
		),
		RunE: c.update,
		Args: cobra.ExactArgs(1),
	}
	updateCmd.Flags().String(configFileFlagName, "", "Name of the file containing link config overrides. "+
		"Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")
	check(updateCmd.MarkFlagRequired(configFileFlagName))
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)
}

func (c *linkCommand) list(cmd *cobra.Command, args []string) error {
	includeTopics, err := cmd.Flags().GetBool(includeTopicsFlagName)
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	listLinksRespDataList, httpResp, err := kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksGet(
		kafkaREST.Context, lkc)
	if err != nil {
		return handleOpenApiError(httpResp, err, kafkaREST)
	}

	if includeTopics {
		outputWriter, err := output.NewListOutputWriter(
			cmd,
			listLinkFieldsIncludeTopics,
			humanListLinkFieldsIncludeTopics,
			structuredListLinkFieldsIncludeTopics)
		if err != nil {
			return err
		}

		for _, link := range listLinksRespDataList.Data {
			if len(link.TopicNames) > 0 {
				for _, topic := range link.TopicNames {
					outputWriter.AddElement(
						&LinkTopicWriter{
							LinkName:        link.LinkName,
							TopicName:       topic,
							SourceClusterId: link.SourceClusterId,
						})
				}
			} else {
				outputWriter.AddElement(
					&LinkTopicWriter{
						LinkName:        link.LinkName,
						TopicName:       "",
						SourceClusterId: link.SourceClusterId,
					})
			}
		}

		return outputWriter.Out()
	} else {
		outputWriter, err := output.NewListOutputWriter(
			cmd, listLinkFields, humanListLinkFields, structuredListLinkFields)
		if err != nil {
			return err
		}

		for _, link := range listLinksRespDataList.Data {
			outputWriter.AddElement(&LinkTopicWriter{
				LinkName:        link.LinkName,
				SourceClusterId: link.SourceClusterId,
			})
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

	sourceClusterId, err := cmd.Flags().GetString(sourceClusterIdFlagName)
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

	configFile, err := cmd.Flags().GetString(configFileFlagName)
	if err != nil {
		return err
	}

	configMap, err := utils.ReadConfigsFromFile(configFile)
	if err != nil {
		return err
	}

	// Two optional flags: --source-api-key and --source-api-secret
	// 1. if I have neither flag set, then no change in behavior – use config-file as normal
	//
	// 2. if I have only 1 flag set, but not the other, then throw an error
	//
	// 3. if I have both set, then the CLI should add these configs on top of configs passed in config-file
	apiKey, err := cmd.Flags().GetString(apiKeyFlagName)
	if err != nil {
		return err
	}

	apiSecret, err := cmd.Flags().GetString(apiSecretFlagName)
	if err != nil {
		return err
	}

	// Overriding the security props by the flag value
	if apiKey != "" && apiSecret != "" {
		configMap[securityProtocalPropertyName] = "SASL_SSL"
		configMap[saslMechanismPropertyName] = "PLAIN"
		configMap[saslJaasConfigPropertyName] = fmt.Sprintf(
			"org.apache.kafka.common.security.plain.PlainLoginModule required "+
				"username=\"%s\" "+
				"password=\"%s\";", apiKey, apiSecret)
	} else if apiKey != "" {
		return errors.New("--source-api-key and --source-api-secret must be used together. " +
			"You cannot pass in one without the other.")
	}

	// Overriding the bootstrap server prop by the flag value
	configMap[sourceBootstrapServersPropertyName] = bootstrapServers

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	createLinkOpt := &kafkarestv3.ClustersClusterIdLinksPostOpts{
		ValidateOnly: optional.NewBool(validateOnly),
		ValidateLink: optional.NewBool(!skipValidatingLink),
		CreateLinkRequestData: optional.NewInterface(kafkarestv3.CreateLinkRequestData{
			SourceClusterId: sourceClusterId,
			Configs:         toCreateTopicConfigs(configMap),
		}),
	}

	httpResp, err := kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksPost(
		kafkaREST.Context, lkc, linkName, createLinkOpt)

	if err == nil {
		msg := errors.CreatedLinkMsg
		if validateOnly {
			msg = errors.DryRunPrefix + msg
		}
		utils.Printf(cmd, msg, linkName)
		return nil
	}

	return handleOpenApiError(httpResp, err, kafkaREST)
}

func (c *linkCommand) delete(cmd *cobra.Command, args []string) error {
	linkName := args[0]
	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	httpResp, err := kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksLinkNameDelete(kafkaREST.Context, lkc, linkName)
	if err == nil {
		utils.Printf(cmd, errors.DeletedLinkMsg, linkName)
	}

	return handleOpenApiError(httpResp, err, kafkaREST)
}

func (c *linkCommand) describe(cmd *cobra.Command, args []string) error {
	linkName := args[0]
	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	listLinkConfigsRespData, httpResp, err := kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksLinkNameConfigsGet(
		kafkaREST.Context, lkc, linkName)
	if err != nil {
		return handleOpenApiError(httpResp, err, kafkaREST)
	}

	outputWriter, err := output.NewListOutputWriter(
		cmd, describeLinkConfigFields, humanDescribeLinkConfigFields, structuredDescribeLinkConfigFields)
	if err != nil {
		return err
	}

	if len(listLinkConfigsRespData.Data) < 1 {
		return outputWriter.Out()
	}

	outputWriter.AddElement(&LinkConfigWriter{
		ConfigName:  "dest.cluster.id",
		ConfigValue: listLinkConfigsRespData.Data[0].ClusterId,
		ReadOnly:    true,
		Sensitive:   true,
		Source:      "",
		Synonyms:    nil,
	})

	for _, config := range listLinkConfigsRespData.Data {
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

func (c *linkCommand) update(cmd *cobra.Command, args []string) error {
	linkName := args[0]
	configFile, err := cmd.Flags().GetString(configFileFlagName)
	if err != nil {
		return err
	}

	configsMap, err := utils.ReadConfigsFromFile(configFile)
	if err != nil {
		return err
	}

	if len(configsMap) == 0 {
		return errors.New(errors.EmptyConfigErrorMsg)
	}

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand, cmd)
	if err != nil {
		return err
	}

	kafkaRestConfigs := toAlterConfigBatchRequestData(configsMap)

	httpResp, err := kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksLinkNameConfigsalterPut(
		kafkaREST.Context, lkc, linkName,
		&kafkarestv3.ClustersClusterIdLinksLinkNameConfigsalterPutOpts{
			AlterConfigBatchRequestData: optional.NewInterface(
				kafkarestv3.AlterConfigBatchRequestData{Data: kafkaRestConfigs}),
		})
	if err == nil {
		utils.Printf(cmd, errors.UpdatedLinkMsg, linkName)
		return nil
	}

	return handleOpenApiError(httpResp, err, kafkaREST)
}
