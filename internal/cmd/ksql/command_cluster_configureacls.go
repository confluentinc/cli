package ksql

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	ksqlv2 "github.com/confluentinc/ccloud-sdk-go-v2/ksql/v2"

	"github.com/confluentinc/cli/internal/pkg/acl"
	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *ksqlCommand) newConfigureAclsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "configure-acls <id> [topic-1] [topic-2] ... [topic-n]",
		Short:             "Configure ACLs for a ksqlDB cluster.",
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
	// Get the Kafka Cluster
	kafkaCluster, err := dynamicconfig.KafkaCluster(c.Context)
	if err != nil {
		return err
	}

	ksqlCluster := args[0]

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	// Ensure the KSQL cluster talks to the current Kafka Cluster
	cluster, err := c.V2Client.DescribeKsqlCluster(ksqlCluster, environmentId)
	if err != nil {
		return err
	}

	if cluster.Spec.KafkaCluster.Id != kafkaCluster.Id {
		output.ErrPrintf(errors.KsqlDBNotBackedByKafkaMsg, ksqlCluster, cluster.Spec.KafkaCluster.Id, kafkaCluster.Id, cluster.Spec.KafkaCluster.Id)
	}

	credentialIdentity := cluster.Spec.CredentialIdentity.GetId()
	if resource.LookupType(credentialIdentity) != resource.ServiceAccount {
		return fmt.Errorf(errors.KsqlDBNoServiceAccountErrorMsg, ksqlCluster)
	}

	serviceAccountId, err := c.getServiceAccount(&cluster)
	if err != nil {
		return err
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	bindings := buildACLBindings(serviceAccountId, &cluster, args[1:])
	if dryRun {
		return acl.PrintACLs(cmd, bindings)
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	httpResp, err := kafkaREST.CloudClient.BatchCreateKafkaAcls(kafkaClusterConfig.ID, getCreateAclRequestDataList(bindings))
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	return acl.PrintACLs(cmd, bindings)
}

func (c *ksqlCommand) getServiceAccount(cluster *ksqlv2.KsqldbcmV2Cluster) (string, error) {
	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return "", err
	}

	credentialIdentity := cluster.Spec.CredentialIdentity.GetId()

	for _, user := range users {
		if user.ServiceName == fmt.Sprintf("KSQL.%s", cluster.GetId()) || user.ResourceId == credentialIdentity {
			return strconv.Itoa(int(user.Id)), nil
		}
	}
	return "", errors.Errorf(errors.KsqlDBNoServiceAccountErrorMsg, cluster.GetId())
}

func buildACLBindings(serviceAccountId string, cluster *ksqlv2.KsqldbcmV2Cluster, topics []string) []*ccstructs.ACLBinding {
	var bindings []*ccstructs.ACLBinding //nolint:prealloc

	for _, operation := range []ccstructs.ACLOperations_ACLOperation{
		ccstructs.ACLOperations_DESCRIBE,
		ccstructs.ACLOperations_DESCRIBE_CONFIGS,
	} {
		bindings = append(bindings, createACL(ccstructs.ResourceTypes_CLUSTER, "kafka-cluster", ccstructs.PatternTypes_LITERAL, serviceAccountId, operation))
	}

	topicPrefix := cluster.Status.GetTopicPrefix()

	for _, operation := range []ccstructs.ACLOperations_ACLOperation{
		ccstructs.ACLOperations_CREATE,
		ccstructs.ACLOperations_DESCRIBE,
		ccstructs.ACLOperations_ALTER,
		ccstructs.ACLOperations_DESCRIBE_CONFIGS,
		ccstructs.ACLOperations_ALTER_CONFIGS,
		ccstructs.ACLOperations_READ,
		ccstructs.ACLOperations_WRITE,
		ccstructs.ACLOperations_DELETE,
	} {
		bindings = append(bindings, createACL(ccstructs.ResourceTypes_TOPIC, topicPrefix, ccstructs.PatternTypes_PREFIXED, serviceAccountId, operation))
		bindings = append(bindings, createACL(ccstructs.ResourceTypes_TOPIC, "_confluent-ksql-"+topicPrefix, ccstructs.PatternTypes_PREFIXED, serviceAccountId, operation))
		bindings = append(bindings, createACL(ccstructs.ResourceTypes_GROUP, "_confluent-ksql-"+topicPrefix, ccstructs.PatternTypes_PREFIXED, serviceAccountId, operation))
	}

	for _, operation := range []ccstructs.ACLOperations_ACLOperation{
		ccstructs.ACLOperations_DESCRIBE,
		ccstructs.ACLOperations_DESCRIBE_CONFIGS,
	} {
		bindings = append(bindings, createACL(ccstructs.ResourceTypes_TOPIC, "*", ccstructs.PatternTypes_LITERAL, serviceAccountId, operation))
		bindings = append(bindings, createACL(ccstructs.ResourceTypes_GROUP, "*", ccstructs.PatternTypes_LITERAL, serviceAccountId, operation))
	}

	for _, operation := range []ccstructs.ACLOperations_ACLOperation{
		ccstructs.ACLOperations_DESCRIBE,
		ccstructs.ACLOperations_DESCRIBE_CONFIGS,
		ccstructs.ACLOperations_READ,
	} {
		for _, topic := range topics {
			bindings = append(bindings, createACL(ccstructs.ResourceTypes_TOPIC, topic, ccstructs.PatternTypes_LITERAL, serviceAccountId, operation))
		}
	}

	for _, operation := range []ccstructs.ACLOperations_ACLOperation{
		ccstructs.ACLOperations_DESCRIBE,
		ccstructs.ACLOperations_WRITE,
	} {
		bindings = append(bindings, createACL(ccstructs.ResourceTypes_TRANSACTIONAL_ID, topicPrefix, ccstructs.PatternTypes_LITERAL, serviceAccountId, operation))
	}

	return bindings
}

func createACL(resourceType ccstructs.ResourceTypes_ResourceType, name string, patternType ccstructs.PatternTypes_PatternType, serviceAccountId string, operation ccstructs.ACLOperations_ACLOperation) *ccstructs.ACLBinding {
	return &ccstructs.ACLBinding{
		Pattern: &ccstructs.ResourcePatternConfig{
			ResourceType: resourceType,
			Name:         name,
			PatternType:  patternType,
		},
		Entry: &ccstructs.AccessControlEntryConfig{
			Principal:      "User:" + serviceAccountId,
			Operation:      operation,
			Host:           "*",
			PermissionType: ccstructs.ACLPermissionTypes_ALLOW,
		},
	}
}

func getCreateAclRequestDataList(bindings []*ccstructs.ACLBinding) kafkarestv3.CreateAclRequestDataList {
	data := make([]kafkarestv3.CreateAclRequestData, len(bindings))
	for i, binding := range bindings {
		data[i] = acl.GetCreateAclRequestData(binding)
	}
	return kafkarestv3.CreateAclRequestDataList{Data: data}
}
