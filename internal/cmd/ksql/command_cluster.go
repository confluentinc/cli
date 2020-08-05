package ksql

import (
	"context"
	"fmt"
	"strconv"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
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
	*pcmd.AuthenticatedCLICommand
	prerunner pcmd.PreRunner
}

// NewClusterCommand returns the Cobra clusterCommand for Ksql Cluster.
func NewClusterCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "app",
			Short: "Manage ksqlDB apps.",
		}, prerunner)
	cmd := &clusterCommand{AuthenticatedCLICommand: cliCmd}
	cmd.prerunner = prerunner
	cmd.init()
	return cmd.Command
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
	createCmd.Flags().String("cluster", "", "Kafka cluster ID.")
	createCmd.Flags().Int32("csu", 4, "Number of CSUs to use in the cluster.")
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
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

	c.AddCommand(&cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a ksqlDB app.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.delete),
	})

	aclsCmd := &cobra.Command{
		Use:   "configure-acls <id> TOPICS...",
		Short: "Configure ACLs for a ksqlDB cluster.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  pcmd.NewCLIRunE(c.configureACLs),
	}
	aclsCmd.Flags().String("cluster", "", "Kafka cluster ID.")
	aclsCmd.Flags().BoolVar(&aclsDryRun, "dry-run", false, "If specified, print the ACLs that will be set and exit.")
	aclsCmd.Flags().SortFlags = false
	c.AddCommand(aclsCmd)
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
		pcmd.ErrPrintln(cmd, errors.EndPointNotPopulatedMsg)
	}
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
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId(), Id: args[0]}
	err := c.Client.KSQL.Delete(context.Background(), req)
	if err != nil {
		return err
	}
	pcmd.Printf(cmd, errors.KsqlDBDeletedMsg, args[0])
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
		if user.ServiceName == fmt.Sprintf("KSQL.%s", cluster.Id) {
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
		pcmd.ErrPrintf(cmd, errors.KsqlDBNotBackedByKafkaMsg, args[0], cluster.KafkaClusterId, kafkaCluster.Id, cluster.KafkaClusterId)
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
