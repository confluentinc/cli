package ksql

import (
	"context"
	"fmt"
	"strconv"

	ksql "github.com/confluentinc/ccloud-sdk-go-v2-internal/ksql/v2"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *ksqlCommand) newConfigureAclsCommand(resource string) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "configure-acls <id> TOPICS...",
		Short:             fmt.Sprintf("Configure ACLs for a ksqlDB %s.", resource),
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.configureACLs,
	}

	cmd.Flags().Bool("dry-run", false, "If specified, print the ACLs that will be set and exit.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *ksqlCommand) configureACLs(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get the Kafka Cluster
	kafkaCluster, err := dynamicconfig.KafkaCluster(c.Context)
	if err != nil {
		return err
	}

	// Ensure the KSQL cluster talks to the current Kafka Cluster
	clusterId := args[0]
	cluster, err := c.V2Client.DescribeKsqlCluster(clusterId, c.EnvironmentId())
	if err != nil {
		return err
	}

	if ksqlKafkaClusterId := cluster.Spec.KafkaCluster.Id; ksqlKafkaClusterId != kafkaCluster.Id {
		utils.ErrPrintf(cmd, errors.KsqlDBNotBackedByKafkaMsg, clusterId, ksqlKafkaClusterId, kafkaCluster.Id, ksqlKafkaClusterId)
	}

	credentialIdentity := cluster.Spec.GetCredentialIdentity().Id
	fmt.Println("CredentialIdentity", credentialIdentity)
	if resource.LookupType(credentialIdentity) != resource.ServiceAccount {
		return fmt.Errorf(errors.KsqlDBNoServiceAccountErrorMsg, clusterId)
	}

	serviceAccountId, err := c.getServiceAccount(&cluster)
	if err != nil {
		return err
	}

	// Setup ACLs
	aclsDryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	bindings := c.buildACLBindings(serviceAccountId, &cluster, args[1:])
	if aclsDryRun {
		return acl.PrintACLs(cmd, bindings, cmd.OutOrStderr())
	}

	return c.Client.Kafka.CreateACLs(ctx, kafkaCluster, bindings)
}

func (c *ksqlCommand) getServiceAccount(cluster *ksql.KsqldbcmV2Cluster) (string, error) {
	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return "", err
	}

	credentialIdentity := cluster.Spec.GetCredentialIdentity().Id

	fmt.Println("got users", len(users))
	for _, user := range users {
		fmt.Println("User.ResourceId", user.ResourceId)
		if user.ServiceName == fmt.Sprintf("KSQL.%s", *cluster.Id) || user.ResourceId == credentialIdentity {
			return strconv.Itoa(int(user.Id)), nil
		}
	}
	return "", errors.Errorf(errors.KsqlDBNoServiceAccountErrorMsg, *cluster.Id)
}

func (c *ksqlCommand) buildACLBindings(serviceAccountId string, cluster *ksql.KsqldbcmV2Cluster, topics []string) []*schedv1.ACLBinding {
	bindings := make([]*schedv1.ACLBinding, 0)
	for _, op := range []schedv1.ACLOperations_ACLOperation{
		schedv1.ACLOperations_DESCRIBE,
		schedv1.ACLOperations_DESCRIBE_CONFIGS,
	} {
		bindings = append(bindings, c.createClusterAcl(op, serviceAccountId))
	}
	topicPrefix := cluster.Status.GetTopicPrefix()
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
		bindings = append(bindings, c.createACL(topicPrefix, schedv1.PatternTypes_PREFIXED, op, schedv1.ResourceTypes_TOPIC, serviceAccountId))
		bindings = append(bindings, c.createACL("_confluent-ksql-"+topicPrefix, schedv1.PatternTypes_PREFIXED, op, schedv1.ResourceTypes_TOPIC, serviceAccountId))
		bindings = append(bindings, c.createACL("_confluent-ksql-"+topicPrefix, schedv1.PatternTypes_PREFIXED, op, schedv1.ResourceTypes_GROUP, serviceAccountId))
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
		bindings = append(bindings, c.createACL(topicPrefix, schedv1.PatternTypes_LITERAL, op, schedv1.ResourceTypes_TRANSACTIONAL_ID, serviceAccountId))
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
