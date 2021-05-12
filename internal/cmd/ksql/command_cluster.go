package ksql

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	"github.com/confluentinc/cli/internal/pkg/acl"
	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/shell/completer"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	listFields                = []string{"Id", "Name", "OutputTopicPrefix", "KafkaClusterId", "Storage", "Endpoint", "Status"}
	listHumanLabels           = []string{"Id", "Name", "Topic Prefix", "Kafka", "Storage", "Endpoint", "Status"}
	listStructuredLabels      = []string{"id", "name", "topic_prefix", "kafka", "storage", "endpoint", "status"}
	describeFields            = []string{"Id", "Name", "OutputTopicPrefix", "KafkaClusterId", "Storage", "Endpoint", "Status"}
	describeHumanRenames      = map[string]string{"KafkaClusterId": "Kafka", "OutputTopicPrefix": "Topic Prefix"}
	describeStructuredRenames = map[string]string{"KafkaClusterId": "kafka", "OutputTopicPrefix": "topic_prefix"}
	aclsDryRun                = false
)

type clusterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner               pcmd.PreRunner
	completableChildren     []*cobra.Command
	completableFlagChildren map[string][]*cobra.Command
	analyticsClient         analytics.Client
}

// NewClusterCommand returns the Cobra clusterCommand for Ksql Cluster.
func NewClusterCommand(prerunner pcmd.PreRunner, analyticsClient analytics.Client) *clusterCommand {
	cliCmd := pcmd.NewAuthenticatedStateFlagCommand(
		&cobra.Command{
			Use:   "app",
			Short: "Manage ksqlDB apps.",
		}, prerunner, SubcommandFlags)
	cmd := &clusterCommand{AuthenticatedStateFlagCommand: cliCmd, analyticsClient: analyticsClient}
	cmd.prerunner = prerunner
	cmd.init()
	return cmd
}

func (c *clusterCommand) Cmd() *cobra.Command {
	return c.Command
}

func (c *clusterCommand) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId()}
	clusters, err := c.Client.KSQL.List(context.Background(), req)
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

func (c *clusterCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List ksqlDB apps.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a ksqlDB app.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.create),
	}
	createCmd.Flags().Int32("csu", 4, "Number of CSUs to use in the cluster.")
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	createCmd.Flags().String("api-key", "", "Kafka API key for the ksqlDB cluster to use (recommended).")
	createCmd.Flags().String("api-secret", "", "Secret for the Kafka API key (recommended).")
	createCmd.Flags().String("image", "", "Image to run (internal).")
	_ = createCmd.Flags().MarkHidden("image")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a ksqlDB app.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.describe),
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	c.AddCommand(describeCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a ksqlDB app.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.delete),
	}
	c.AddCommand(deleteCmd)

	aclsCmd := &cobra.Command{
		Use:   "configure-acls <id> TOPICS...",
		Short: "Configure ACLs for a ksqlDB cluster.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  pcmd.NewCLIRunE(c.configureACLs),
	}
	aclsCmd.Flags().BoolVar(&aclsDryRun, "dry-run", false, "If specified, print the ACLs that will be set and exit.")
	aclsCmd.Flags().SortFlags = false
	c.AddCommand(aclsCmd)

	c.completableChildren = []*cobra.Command{describeCmd, deleteCmd, aclsCmd}
	c.completableFlagChildren = map[string][]*cobra.Command{
		"cluster": {createCmd},
	}
}

func (c *clusterCommand) list(cmd *cobra.Command, _ []string) error {
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId()}
	clusters, err := c.Client.KSQL.List(context.Background(), req)
	if err != nil {
		return err
	}
	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}
	for _, cluster := range clusters {
		outputWriter.AddElement(cluster)
	}
	return outputWriter.Out()
}

func (c *clusterCommand) create(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return err
	}
	csus, err := cmd.Flags().GetInt32("csu")
	if err != nil {
		return err
	}
	cfg := &schedv1.KSQLClusterConfig{
		AccountId:      c.EnvironmentId(),
		Name:           args[0],
		TotalNumCsu:    uint32(csus),
		KafkaClusterId: kafkaCluster.ID,
	}

	kafkaApiKey, err := cmd.Flags().GetString("api-key")
	if err != nil {
		return err
	}
	kafkaApiKeySecret, err := cmd.Flags().GetString("api-secret")
	if err != nil {
		return err
	}

	if kafkaApiKey != "" && kafkaApiKeySecret != "" {
		cfg.KafkaApiKey = &schedv1.ApiKey{
			Key:    kafkaApiKey,
			Secret: kafkaApiKeySecret,
		}
	} else if (kafkaApiKey == "" && kafkaApiKeySecret != "") || (kafkaApiKeySecret == "" && kafkaApiKey != "") {
		return fmt.Errorf(errors.APIKeyAndSecretBothRequired)
	} else {
		_, _ = fmt.Fprintln(os.Stderr, errors.KSQLCreateDeprecateWarning)
	}

	image, err := cmd.Flags().GetString("image")
	if err == nil && len(image) > 0 {
		cfg.Image = image
	}
	cluster, err := c.Client.KSQL.Create(context.Background(), cfg)
	if err != nil {
		return err
	}
	// use count to prevent the command from hanging too long waiting for the endpoint value
	count := 0
	// endpoint value filled later, loop until endpoint information is not null (usually just one describe call is enough)
	for cluster.Endpoint == "" && count < 3 {
		req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId(), Id: cluster.Id}
		cluster, err = c.Client.KSQL.Describe(context.Background(), req)
		if err != nil {
			return err
		}
		count += 1
	}
	if cluster.Endpoint == "" {
		utils.ErrPrintln(cmd, errors.EndPointNotPopulatedMsg)
	}
	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, cluster.Id)
	return output.DescribeObject(cmd, cluster, describeFields, describeHumanRenames, describeStructuredRenames)
}

func (c *clusterCommand) describe(cmd *cobra.Command, args []string) error {
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId(), Id: args[0]}
	cluster, err := c.Client.KSQL.Describe(context.Background(), req)
	if err != nil {
		err = errors.CatchKSQLNotFoundError(err, args[0])
		return err
	}
	return output.DescribeObject(cmd, cluster, describeFields, describeHumanRenames, describeStructuredRenames)
}

func (c *clusterCommand) delete(cmd *cobra.Command, args []string) error {
	id := args[0]
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId(), Id: id}
	err := c.Client.KSQL.Delete(context.Background(), req)
	if err != nil {
		return err
	}
	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, id)
	utils.Printf(cmd, errors.KsqlDBDeletedMsg, args[0])
	return nil
}

func (c *clusterCommand) createACL(prefix string, patternType schedv1.PatternTypes_PatternType, operation schedv1.ACLOperations_ACLOperation, resource schedv1.ResourceTypes_ResourceType, serviceAccountId string) *schedv1.ACLBinding {
	binding := &schedv1.ACLBinding{
		Entry: &schedv1.AccessControlEntryConfig{
			Host: "*",
		},
		Pattern: &schedv1.ResourcePatternConfig{},
	}
	binding.Entry.PermissionType = schedv1.ACLPermissionTypes_ALLOW
	binding.Entry.Operation = operation
	binding.Entry.Principal = "User:" + serviceAccountId
	binding.Pattern.PatternType = patternType
	binding.Pattern.ResourceType = resource
	binding.Pattern.Name = prefix
	return binding
}

func (c *clusterCommand) createClusterAcl(operation schedv1.ACLOperations_ACLOperation, serviceAccountId string) *schedv1.ACLBinding {
	binding := &schedv1.ACLBinding{
		Entry: &schedv1.AccessControlEntryConfig{
			Host: "*",
		},
		Pattern: &schedv1.ResourcePatternConfig{},
	}
	binding.Entry.PermissionType = schedv1.ACLPermissionTypes_ALLOW
	binding.Entry.Operation = operation
	binding.Entry.Principal = "User:" + serviceAccountId
	binding.Pattern.PatternType = schedv1.PatternTypes_LITERAL
	binding.Pattern.ResourceType = schedv1.ResourceTypes_CLUSTER
	binding.Pattern.Name = "kafka-cluster"
	return binding
}

func (c *clusterCommand) buildACLBindings(serviceAccountId string, cluster *schedv1.KSQLCluster, topics []string) []*schedv1.ACLBinding {
	bindings := make([]*schedv1.ACLBinding, 0)
	for _, op := range []schedv1.ACLOperations_ACLOperation{
		schedv1.ACLOperations_DESCRIBE,
		schedv1.ACLOperations_DESCRIBE_CONFIGS,
	} {
		bindings = append(bindings, c.createClusterAcl(op, serviceAccountId))
	}
	for _, op := range []schedv1.ACLOperations_ACLOperation{
		schedv1.ACLOperations_CREATE,
		schedv1.ACLOperations_DESCRIBE,
		schedv1.ACLOperations_ALTER,
		schedv1.ACLOperations_DESCRIBE_CONFIGS,
		schedv1.ACLOperations_ALTER_CONFIGS,
		schedv1.ACLOperations_READ,
		schedv1.ACLOperations_WRITE,
		schedv1.ACLOperations_DELETE,
	} {
		bindings = append(bindings, c.createACL(cluster.OutputTopicPrefix, schedv1.PatternTypes_PREFIXED, op, schedv1.ResourceTypes_TOPIC, serviceAccountId))
		bindings = append(bindings, c.createACL("_confluent-ksql-"+cluster.OutputTopicPrefix, schedv1.PatternTypes_PREFIXED, op, schedv1.ResourceTypes_TOPIC, serviceAccountId))
		bindings = append(bindings, c.createACL("_confluent-ksql-"+cluster.OutputTopicPrefix, schedv1.PatternTypes_PREFIXED, op, schedv1.ResourceTypes_GROUP, serviceAccountId))
	}
	for _, op := range []schedv1.ACLOperations_ACLOperation{
		schedv1.ACLOperations_DESCRIBE,
		schedv1.ACLOperations_DESCRIBE_CONFIGS,
	} {
		bindings = append(bindings, c.createACL("*", schedv1.PatternTypes_LITERAL, op, schedv1.ResourceTypes_TOPIC, serviceAccountId))
		bindings = append(bindings, c.createACL("*", schedv1.PatternTypes_LITERAL, op, schedv1.ResourceTypes_GROUP, serviceAccountId))
	}
	for _, op := range []schedv1.ACLOperations_ACLOperation{
		schedv1.ACLOperations_DESCRIBE,
		schedv1.ACLOperations_DESCRIBE_CONFIGS,
		schedv1.ACLOperations_READ,
	} {
		for _, t := range topics {
			bindings = append(bindings, c.createACL(t, schedv1.PatternTypes_LITERAL, op, schedv1.ResourceTypes_TOPIC, serviceAccountId))
		}
	}
	// for transactional produces to command topic
	for _, op := range []schedv1.ACLOperations_ACLOperation{
		schedv1.ACLOperations_DESCRIBE,
		schedv1.ACLOperations_WRITE,
	} {
		bindings = append(bindings, c.createACL(cluster.PhysicalClusterId, schedv1.PatternTypes_LITERAL, op, schedv1.ResourceTypes_TRANSACTIONAL_ID, serviceAccountId))
	}
	return bindings
}

func (c *clusterCommand) getServiceAccount(cluster *schedv1.KSQLCluster) (string, error) {
	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return "", err
	}

	for _, user := range users {
		if user.ServiceName == fmt.Sprintf("KSQL.%s", cluster.Id) || (cluster.KafkaApiKey != nil && user.Id == cluster.KafkaApiKey.UserId) {
			return strconv.Itoa(int(user.Id)), nil
		}
	}
	return "", errors.Errorf(errors.NoServiceAccountErrorMsg, cluster.Id)
}

func (c *clusterCommand) configureACLs(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get the Kafka Cluster
	kafkaCluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}

	// Ensure the KSQL cluster talks to the current Kafka Cluster
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId(), Id: args[0]}
	cluster, err := c.Client.KSQL.Describe(context.Background(), req)
	if err != nil {
		return err
	}
	if cluster.KafkaClusterId != kafkaCluster.Id {
		utils.ErrPrintf(cmd, errors.KsqlDBNotBackedByKafkaMsg, args[0], cluster.KafkaClusterId, kafkaCluster.Id, cluster.KafkaClusterId)
	}

	if cluster.ServiceAccountId == 0 {
		return fmt.Errorf(errors.KsqlDBNoServiceAccount, args[0])
	}


	serviceAccountId, err := c.getServiceAccount(cluster)
	if err != nil {
		return err
	}

	// Setup ACLs
	bindings := c.buildACLBindings(serviceAccountId, cluster, args[1:])
	if aclsDryRun {
		return acl.PrintACLs(cmd, bindings, cmd.OutOrStderr())
	}
	err = c.Client.Kafka.CreateACLs(ctx, kafkaCluster, bindings)
	if err != nil {
		return err
	}
	return nil
}

func (c *clusterCommand) ServerCompletableFlagChildren() map[string][]*cobra.Command {
	return c.completableFlagChildren
}

func (c *clusterCommand) ServerFlagComplete() map[string]func() []prompt.Suggest {
	return map[string]func() []prompt.Suggest{
		"cluster": completer.ClusterFlagServerCompleterFunc(c.Client, c.EnvironmentId()),
	}
}
