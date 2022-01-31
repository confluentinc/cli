package ksql

import (
	"context"
	"fmt"
	"os"
	"strconv"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *ksqlCommand) newConfigureAclsCommand(isApp bool) *cobra.Command {
	shortText := "Configure ACLs for a ksqlDB cluster."
	longText := ""
	runCommand := c.configureACLsCluster
	if isApp {
		// DEPRECATED: this should be removed before CLI v3, this work is tracked in https://confluentinc.atlassian.net/browse/KCI-1411
		shortText = "(Deprecated) Configure ACLs for a ksqlDB app."
		longText = "Configure ACLs for a ksqlDB app. " + errors.KSQLAppDeprecateWarning
		runCommand = c.configureACLsApp
	}

	cmd := &cobra.Command{
		Use:               "configure-acls <id> TOPICS...",
		Short:             shortText,
		Long:              longText,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              pcmd.NewCLIRunE(runCommand),
	}

	cmd.Flags().Bool("dry-run", false, "If specified, print the ACLs that will be set and exit.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *ksqlCommand) configureACLsCluster(cmd *cobra.Command, args []string) error {
	return c.configureACLs(cmd, args, false)
}

func (c *ksqlCommand) configureACLsApp(cmd *cobra.Command, args []string) error {
	return c.configureACLs(cmd, args, true)
}

func (c *ksqlCommand) configureACLs(cmd *cobra.Command, args []string, isApp bool) error {
	ctx := context.Background()

	// Get the Kafka Cluster
	kafkaCluster, err := pcmd.KafkaCluster(c.Context)
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
		return fmt.Errorf(errors.KsqlDBNoServiceAccountErrorMsg, args[0])
	}

	serviceAccountId, err := c.getServiceAccount(cluster)
	if err != nil {
		return err
	}

	// Setup ACLs
	aclsDryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	bindings := c.buildACLBindings(serviceAccountId, cluster, args[1:])
	if aclsDryRun {
		return acl.PrintACLs(cmd, bindings, cmd.OutOrStderr())
	}

	if isApp {
		_, _ = fmt.Fprintln(os.Stderr, errors.KSQLAppDeprecateWarning)
	}
	return c.Client.Kafka.CreateACLs(ctx, kafkaCluster, bindings)
}

func (c *ksqlCommand) getServiceAccount(cluster *schedv1.KSQLCluster) (string, error) {
	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return "", err
	}

	for _, user := range users {
		if user.ServiceName == fmt.Sprintf("KSQL.%s", cluster.Id) || (cluster.KafkaApiKey != nil && user.Id == cluster.KafkaApiKey.UserId) {
			return strconv.Itoa(int(user.Id)), nil
		}
	}
	return "", errors.Errorf(errors.KsqlDBNoServiceAccountErrorMsg, cluster.Id)
}

func (c *ksqlCommand) buildACLBindings(serviceAccountId string, cluster *schedv1.KSQLCluster, topics []string) []*schedv1.ACLBinding {
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

func (c *ksqlCommand) createClusterAcl(operation schedv1.ACLOperations_ACLOperation, serviceAccountId string) *schedv1.ACLBinding {
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

func (c *ksqlCommand) createACL(prefix string, patternType schedv1.PatternTypes_PatternType, operation schedv1.ACLOperations_ACLOperation, resource schedv1.ResourceTypes_ResourceType, serviceAccountId string) *schedv1.ACLBinding {
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
